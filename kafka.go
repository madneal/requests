package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"reflect"
)

type Request struct {
	Url       string
	Headers   map[string]string
	Method    string
	Host      string
	AgentId   string
	Timestamp int64
	Postdata  string
}

var zeekMsg = [...]string{"Content-Type", "Accept-Encoding", "Referer", "Cookie", "Origin", "Host", "Accept-Language",
	"Accept", "Accept-Charset", "Connection", "User-Agent"}

func ReadKafka(topic string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{"localhost:9092"},
		Topic:     topic,
		Partition: 0,
		MinBytes:  1,    // 10KB
		MaxBytes:  1000, // 10MB
	})
	//r.SetOffset(42)

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	defer r.Close()
}

func ParseJson(msg string) Request {
	var request Request
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(msg), &data); err != nil {
		fmt.Println(err)
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
	return request
}
