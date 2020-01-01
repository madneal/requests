package main

import (
	"context"
	md5 "crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"net/url"
	"reflect"
)

var zeekMsg = [...]string{"Content-Type", "Accept-Encoding", "Referer", "Cookie", "Origin", "Host", "Accept-Language",
	"Accept", "Accept-Charset", "Connection", "User-Agent"}

func ReadKafka(topic string, hosts []string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  hosts,
		Topic:    topic,
		GroupID:  "consumer-group-pvs",
		MinBytes: 1,
		MaxBytes: 1000,
	})

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			fmt.Println(err)
			break
		}
		request, err := ParseJson(string(m.Value))
		if err != nil {
			Log.Error(err)
		} else {
			go SendRequest(request)
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	defer r.Close()
}

func ParseJson(msg string) (Request, error) {
	var request Request
	var data map[string]interface{}
	var err error
	if err = json.Unmarshal([]byte(msg), &data); err != nil {
		fmt.Println(err)
		return request, err
	}
	var headersType string
	if _, ok := data["headers"]; ok {
		//fmt.Println(val)
		headersType = reflect.TypeOf(data["headers"]).String()
	}

	request.Host = data["host"].(string)
	request.AgentId = data["agentId"].(string)
	request.Timestamp = int64(data["t"].(float64))
	request.Method = data["method"].(string)
	headers := make(map[string]string)
	// headers is array
	if headersType == "[]interface {}" {
		headers1 := data["headers"].([]interface{})
		for _, header := range headers1 {
			headerMap := header.(map[string]interface{})
			headers[headerMap["Name"].(string)] = headerMap["value"].(string)
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
		request.Url = headers["host"] + data["uri"].(string)
	}
	if request.Url == "" {
		request.Url = data["url"].(string)
	}

	if request.Method == "POST" && data["postdata"].(string) != "" {
		body, err := base64.StdEncoding.DecodeString(data["postdata"].(string))
		if err != nil {
			fmt.Println(err)
		}
		request.Postdata = string(body)
	}
	request.Headers = headers
	return request, err
}

func InsertAsset(request Request) {
	u, err := url.Parse(request.Url)
	if err != nil {
		fmt.Println(err)
	}
	str := fmt.Sprintf("%s%s%s", u.Scheme, u.Host, u.Path)
	md5Str := fmt.Sprintf("%x", md5.Sum([]byte(str)))
	fmt.Println(md5Str)
	asset := Asset{
		Url:    request.Url,
		Method: request.Method,
		Md5:    md5Str,
	}
	exists := Exists(md5Str)
	if !exists {
		NewAsset(asset)
	} else {
		fmt.Println("the record has exists!")
	}
}
