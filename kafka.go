package main

import (
	"context"
	"encoding/json"
	"fmt"
	kafka "github.com/segmentio/kafka-go"
	"reflect"
)

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

func ParseJson(msg string) {
	//var data interface{}

	var request map[string]interface{}
	json.Unmarshal([]byte(msg), &request)
	headersType := request["headers"]
	if headersType == "[]interface {}" {

	} else if headersType == "map[string]interface {}" {

	} else {

	}
	fmt.Println(request["agentId"])
	fmt.Println(request["headers"])
	fmt.Println(reflect.TypeOf(request["headers"]))
	fmt.Println(reflect.TypeOf(request["headers"]).String() == "[]interface {}")
}
