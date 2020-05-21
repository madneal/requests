package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/segmentio/kafka-go"
	"net/url"
	"reflect"
	"strings"
	"time"
)

var zeekMsg = [...]string{"Content-Type", "Accept-Encoding", "Referer", "Cookie", "Origin", "Host", "Accept-Language",
	"Accept", "Accept-Charset", "Connection", "User-Agent"}
var rdb *redis.Client

func ReadKafka(topic string, hosts []string, groupId string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  hosts,
		Topic:    topic,
		GroupID:  groupId,
		MinBytes: 1,
		MaxBytes: 1000,
	})

	//messages := make([]string, 0)

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

func RunTask(msg string) {
	//fmt.Printf("process msg: %s\n", msg)
	request, err := ParseJson(msg)
	if err != nil {
		Log.Error(err)
		fmt.Println("parse request failed")
		return
	} else {
		if CONFIG.Run.Redis == true {
			if rdb.SIsMember(CONFIG.Redis.Set, request.Url).Val() == true {
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
		port := data["resp_p"].(string)
		var schema string
		if port == "443" {
			schema = "https://"
		} else if port == "-" {
			return request, nil
		} else {
			schema = "http://"
		}
		request.Url = schema + headers["Host"] + data["uri"].(string)
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
	asset := CreateAssetByUrl(request.Url)
	if asset == nil {
		return
	}
	asset.Method = request.Method
	err := NewAsset(asset)
	if err != nil {
		Log.Error(err)
	}
}

func CreateAssetByUrl(urlStr string) *Asset {
	u, err := url.Parse(urlStr)
	if err != nil {
		Log.Error(err)
		return nil
	}
	params := ObtainQueryKeys(u)
	return &Asset{
		Url:         fmt.Sprintf("%s%s%s%s", u.Scheme, "://", u.Host, u.Path),
		Params:      params,
		Host:        u.Host,
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
