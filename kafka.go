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

var zeekMsg = [...]string{"Content-Type", "Accept-Encoding", "Referer", "Cookie", "Origin", "Host", "Accept-Language",
	"Accept", "Accept-Charset", "Connection", "User-Agent"}
var rdb *redis.Client

func ReadKafka() {
	if CONFIG.Run.MultiThread {
		fmt.Println("Read kafka as multi thread")
		MultiThreadKafka()
	} else {
		fmt.Println("Read kafka as single thread")
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

		if CONFIG.Run.Debug == true {
			fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
		}
		//var i int

		//if len(messages) <= CONFIG.Run.Threads {
		//	messages = append(messages, string(m.Value))
		//	continue
		//}
		//
		//var wg sync.WaitGroup
		//for j := 0; j < CONFIG.Run.Threads; j++ {
		//	//fmt.Println("Main:Starting worker")
		//	wg.Add(1)
		//	go func(msg string) {
		//		//fmt.Printf("Worker %v: Started\n", j)
		//		RunTask(msg)
		//		wg.Done()
		//		//fmt.Printf("Worker %v: Finished\n", j)
		//	}(messages[j])
		//	wg.Wait()
		//}
		//messages = nil

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

		InsertAsset(request)
		if CONFIG.Run.Production {
			return
		}
		// obtain scheme from referer and send request
		isValidReferer, scheme := IsValidReferer(request)
		if isValidReferer == true {
			url, err := SetUrlByScheme(scheme, request.Url)
			if err != nil {
				Log.Errorf("obtain url for %s by referer failed", request.Url)
			} else {
				request.Url = url
				SendRequest(request)
			}
			return
		}
		SendRequest(request)
		// repeat the request, for http and https respectively
		if strings.Contains(request.Url, "https") {
			request.Url = strings.Replace(request.Url, "https", "http", 1)
		} else {
			request.Url = strings.Replace(request.Url, "http", "https", 1)
		}
		SendRequest(request)
	}
}

func SetUrlByScheme(scheme, urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if nil != err {
		return "", err
	}
	u.Scheme = scheme
	return u.String(), err
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
	var port string
	if data["resp_p"] != nil {
		port = data["resp_p"].(string)
		request.Port, _ = strconv.Atoi(port)
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
		//port := data["resp_p"].(string)
		var schema string
		if port == "-" {
			return request, errors.New(fmt.Sprintf("thr port is -, msg: %s", msg))
		} else {
			schema = "http://"
		}
		request.Url = schema + headers["Host"] + data["uri"].(string)
	}
	if !ValidateHost(request.Host) {
		return request, errors.New(fmt.Sprintf("The host is invalid, msg: %s", msg))
	}
	if request.Url == "" {
		request.Url = data["url"].(string)
	}
	// todo there is not post asset handle for post now
	if !CONFIG.Run.Production && request.Method == "POST" && data["postdata"].(string) != "" {
		body, err := base64.StdEncoding.DecodeString(data["postdata"].(string))
		if err != nil {
			Log.Error(err)
		} else {
			request.Postdata = string(body)
		}
	}
	request.Headers = headers
	return request, err
}

func ValidateHost(host string) bool {
	host = strings.ToLower(host)
	isIp, err := regexp.MatchString(`^\d{1,3}\.`, host)
	if !strings.Contains(host, ".") {
		return false
	}
	if strings.HasPrefix(host, "10.") || strings.HasPrefix(host, "172.") || strings.HasPrefix(host, "127.") {
		return false
	}
	if !(strings.HasSuffix(host, ".com") || strings.HasSuffix(host, ".cn")) && !isIp {
		return false
	}
	matched, err := regexp.MatchString(`(^(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$)|(^(([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9])\.)*([A-Za-z0-9]|[A-Za-z0-9][A-Za-z0-9\-]*[A-Za-z0-9])$)`, host)
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
	return "http://" + host + uri
}

func InsertAsset(request Request) {
	asset := CreateAssetByUrl(request.Url, request.Host)
	asset.Port = request.Port
	if asset == nil {
		return
	}
	asset.Method = request.Method
	asset.Md5 = ComputeHash(asset.Url + asset.Method)
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

func CreateAssetByUrl(urlStr, host string) *Asset {
	u, err := url.Parse(urlStr)
	if err != nil {
		Log.Error(err)
		return nil
	}
	params := ObtainQueryKeys(u)
	return &Asset{
		Url:         fmt.Sprintf("%s%s%s%s", u.Scheme, "://", u.Host, u.Path),
		Params:      params,
		Host:        host,
		Ip:          ObtainIp(u.Host),
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
}

// ObtainQueryKeys is utilized to obtain query keys
func ObtainQueryKeys(u *url.URL) string {
	q := u.Query()
	var result string
	for k, _ := range q {
		result += k + ","
	}
	return result
}

func ObtainIp(host string) string {
	ip := QueryIp(host)
	if ip == "" {
		ips := GetIp(host)
		for _, ipEle := range ips {
			ip += ipEle.String() + ","
		}
	}
	return ip
}
