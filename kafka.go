package main

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/segmentio/kafka-go"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var rdb *redis.Client

func ReadKafka() {
	if CONFIG.Run.MultiThread {
		Log.Info("Read kafka as multi thread")
		MultiThreadKafka()
	} else {
		Log.Info("Read kafka as single thread")
		SingleThreadKafka()
	}
}

func SingleThreadKafka() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  CONFIG.Kafka.Brokers,
		Topic:    CONFIG.Kafka.Topic,
		GroupID:  CONFIG.Kafka.GroupId,
		MinBytes: CONFIG.Kafka.Min,
		MaxBytes: CONFIG.Kafka.Max,
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			Log.Error(err)
			break
		}

		if CONFIG.Run.Debug {
			fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
		}
		RunTask(string(m.Value))
	}
}

func MultiThreadKafka() {
	group, err := kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		ID:      CONFIG.Kafka.GroupId,
		Brokers: CONFIG.Kafka.Brokers,
		Topics:  []string{CONFIG.Kafka.Topic},
	})

	if err != nil {
		Log.Errorf("error creating consumer group: %+v\n", err)
		os.Exit(1)
	}
	defer group.Close()

	for {
		gen, err := group.Next(context.TODO())
		if err != nil {
			Log.Error(err)
			break
		}

		assignments := gen.Assignments[CONFIG.Kafka.Topic]
		for _, assignment := range assignments {
			partition, offset := assignment.ID, assignment.Offset
			gen.Start(func(ctx context.Context) {
				// create reader for this partition.
				reader := kafka.NewReader(kafka.ReaderConfig{
					Brokers:   CONFIG.Kafka.Brokers,
					Topic:     CONFIG.Kafka.Topic,
					Partition: partition,
				})
				defer reader.Close()

				// seek to the last committed offset for this partition.
				reader.SetOffset(offset)
				for {
					msg, err := reader.ReadMessage(ctx)
					switch err {
					case kafka.ErrGenerationEnded:
						// generation has ended.  commit offsets.  in a real app,
						// offsets would be committed periodically.
						gen.CommitOffsets(map[string]map[int]int64{CONFIG.Kafka.Topic: {partition: offset}})
						return
					case nil:
						if CONFIG.Run.Debug {
							fmt.Printf("message at offset %d: %s = %s\n", msg.Offset, string(msg.Key), string(msg.Value))
							Log.Infof("The current partition is %d, and the offset is %d", partition, msg.Offset)
						}
						gen.CommitOffsets(map[string]map[int]int64{CONFIG.Kafka.Topic: {partition: offset}})
						RunTask(string(msg.Value))
						offset = msg.Offset
					default:
						Log.Errorf("error reading message: %+v\n", err)
					}
				}
			})
		}
	}
}

func RunTask(msg string) {
	request, err := ParseJson(msg)
	if err != nil {
		Log.Error(err)
		return
	} else {
		if CONFIG.Run.Redis {
			if rdb.SIsMember(CONFIG.Redis.Set, request.Url).Val() {
				return
			}
			err = rdb.SAdd(CONFIG.Redis.Set, request.Url).Err()
			if err != nil {
				Log.Error(err)
			}
		}

		if CONFIG.Run.Asset {
			InsertAsset(request)
		}

		if CONFIG.Run.Plugin {
			CheckVulns(&request)
		}
	}
}

func ParseJson(msg string) (Request, error) {
	var request Request
	var data map[string]interface{}
	var err error
	if err = json.Unmarshal([]byte(msg), &data); err != nil {
		Log.Error(err)
		return request, err
	}
	var headersType string
	if _, ok := data["headers"]; ok {
		headersType = reflect.TypeOf(data["headers"]).String()
	}
	if data["agentId"] != nil {
		request.AgentId = data["agentId"].(string)
	}
	if data["t"] != nil {
		request.Timestamp = int64(data["t"].(float64))
	}
	if data["method"] != nil {
		request.Method = data["method"].(string)
	}
	if data["status_code"] != nil {
		request.StatusCode = int(data["status_code"].(float64))
		if request.StatusCode == 0 {
			Log.Infof("The request status code is 0, msg: %s", msg)
		}
	}
	if CONFIG.Run.Production && data["id.resp_p"] != nil {
		request.Port = int(data["id.resp_p"].(float64))
	}
	if !CONFIG.Run.Production && data["resp_p"] != nil {
		request.Port, _ = strconv.Atoi(data["resp_p"].(string))
	}
	if request.Port == 0 {
		Log.Warnf("The request port is 0, the msg is %s", msg)
	}
	headers := make(map[string]string)
	// headers is array
	if headersType == "[]interface {}" {
		headers1 := data["headers"].([]interface{})
		for _, header := range headers1 {
			headerMap := header.(map[string]interface{})
			headers[headerMap["name"].(string)] = headerMap["value"].(string)
		}
	} else if headersType == "map[string]interface {}" {
		headers1 := data["headers"].(map[string]interface{})
		for k, v := range headers1 {
			headers[k] = v.(string)
		}
	} else if data["agentId"] == nil {
		if data["host"] != nil {
			request.Host = data["host"].(string)
		} else {
			return request, errors.New(fmt.Sprintf("There is no host in msg, msg: %s", msg))
		}
		request.Url = ObtainUrl(data)
	} else {
		request.Host = data["Host"].(string)
		for _, msg := range zeekMsg {
			if data[msg] == nil {
				continue
			}
			if data[msg].(string) != "-" {
				headers[msg] = data[msg].(string)
			}
			if msg == "User-Agent" {
				headers[msg] = UA
			}
		}
		schema := HTTP_SCHEMA
		request.Url = schema + headers["Host"] + data["uri"].(string)
	}
	if CONFIG.Run.Asset && !ValidateHost(request.Host) {
		return request, errors.New(fmt.Sprintf("The host is invalid, msg: %s", msg))
	}
	if request.Url == "" {
		request.Url = data["url"].(string)
	}

	if !CONFIG.Run.Production && request.Method == POST_METHOD && data["postdata"] != nil {
		if CONFIG.Run.Production {
			request.Postdata = data["postdata"].(string)
		} else {
			body, err := base64.StdEncoding.DecodeString(data["postdata"].(string))
			if err != nil {
				Log.Error(err)
			} else {
				request.Postdata = string(body)
			}
		}
	}
	request.Headers = headers
	return request, err
}

func ValidateHost(host string) bool {
	host = strings.ToLower(host)
	isIp, err := regexp.MatchString(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`, host)
	if !strings.Contains(host, ".") || strings.Contains(host, "*") {
		return false
	}
	if strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "172.") ||
		strings.HasPrefix(host, "127.") || strings.HasPrefix(host, "29.") {
		return false
	}
	hasPort, err := regexp.MatchString(`\w+:\d+`, host)
	if !(strings.HasSuffix(host, ".com") || strings.HasSuffix(host, ".cn") || hasPort) && !isIp {
		return false
	}
	matched, err := regexp.MatchString(`^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])(:\d+)*`, host)
	if err != nil {
		Log.Error(err)
	}
	return matched
}

// ObtainUrl is utilized to obtain url from data
func ObtainUrl(data map[string]interface{}) string {
	var host string
	var uri string
	if data["host"] != nil {
		host = data["host"].(string)
	}
	if data["uri"] != nil {
		uri = data["uri"].(string)
	}
	return HTTP_SCHEMA + host + uri
}

func InsertAsset(request Request) {
	asset := CreateAsset(request.Host, request.Port)
	if CONFIG.Run.Env == QA_ENV {
		asset.Ip = GetIpStr(asset.Host)
	}
	if asset == nil {
		return
	}
	err := NewAsset(asset)
	if err != nil {
		Log.Error(err)
	}
}

func ComputeHash(urlAndMethod string) string {
	h := md5.New()
	h.Write([]byte(urlAndMethod))
	result := hex.EncodeToString(h.Sum(nil))
	return result
}

func CheckIfBlackExtension(url string) bool {
	lowerUrl := strings.ToLower(url)
	for _, extension := range BLACK_EXTENSIONS {
		if strings.HasSuffix(lowerUrl, extension) {
			return true
		}
	}
	return false
}

func CreateAsset(host string, port int) *Asset {
	return &Asset{
		Host:        host,
		Ip:          "",
		Port:        port,
		Env:         CONFIG.Run.Env,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
}

// ObtainQueryKeys is utilized to obtain query keys
func ObtainQueryKeys(u *url.URL) string {
	q := u.Query()
	var result string
	for k := range q {
		result += k + ","
	}
	return result
}
