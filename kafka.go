package main

import (
	"context"
	md5 "crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	kafka "github.com/segmentio/kafka-go"
	"net/url"
	"reflect"
	"strings"
)

var zeekMsg = [...]string{"Content-Type", "Accept-Encoding", "Referer", "Cookie", "Origin", "Host", "Accept-Language",
	"Accept", "Accept-Charset", "Connection", "User-Agent"}
var rdb *redis.Client

func ReadKafka(topic string, hosts []string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  hosts,
		Topic:    topic,
		GroupID:  "consumer-group-pvs",
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
		//fmt.Printf("handle for request %s\n", request.Url)
		if CONFIG.Run.Redis == true {
			//fmt.Println(rdb)
			//fmt.Println(rdb.SIsMember(CONFIG.Redis.Set, request.Url))
			if rdb.SIsMember(CONFIG.Redis.Set, request.Url).Val() == true {
				return
			}
			err = rdb.SAdd(CONFIG.Redis.Set, request.Url).Err()
			if err != nil {
				fmt.Println(err)
			}
		}

		//if len(request.Method) > 5 {
		//	Log.Warnf("The method %s is too long", request.Method)
		//}
		InsertAsset(request)
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
		//fmt.Println(val)
		headersType = reflect.TypeOf(data["headers"]).String()
	}

	request.Host = data["Host"].(string)
	request.AgentId = data["agentId"].(string)
	request.Timestamp = int64(data["t"].(float64))
	request.Method = data["method"].(string)
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
	} else {
		for _, msg := range zeekMsg {
			//fmt.Println(msg)
			if data[msg].(string) != "-" {
				headers[msg] = data[msg].(string)
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
		//fmt.Println(request.Url)
	}
	if request.Url == "" {
		request.Url = data["url"].(string)
	}

	if request.Method == "POST" && data["postdata"].(string) != "" {
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

func InsertAsset(request Request) {
	u, err := url.Parse(request.Url)
	if err != nil {
		Log.Error(err)
		return
	}
	str := fmt.Sprintf("%s%s%s", u.Scheme, u.Host, u.Path)
	md5Str := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	fmt.Println(md5Str)
	asset := Asset{
		Url:    request.Url,
		Method: request.Method,
		Md5:    md5Str,
	}
	err = NewAsset(asset)
	if err != nil {
		Log.Error(err)
	}
}
