package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

func init() {
	source, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		Log.Error(err)
	}

	err = yaml.Unmarshal(source, &CONFIG)
	if err != nil {
		Log.Error(err)
	}
	//fmt.Println(CONFIG.Kafka)
}

func main() {
	if CONFIG.Kafka.Topic == "" {
		fmt.Println("Please set the topic")
	}
	if len(CONFIG.Kafka.Brokers) == 0 {
		fmt.Println("Please set the brokers")
	}
	fmt.Println("*********Begin the Assets detect*************")
	fmt.Printf("kafka topic:%s\n", CONFIG.Kafka.Topic)
	SetDownloadService()
	ReadKafka(CONFIG.Kafka.Topic, CONFIG.Kafka.Brokers, CONFIG.Kafka.GroupId)
}
