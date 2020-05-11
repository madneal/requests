package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
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
	fmt.Println("*********Begin the Assets detect*************")
	if len(os.Args) < 2 {
		fmt.Println("Please speficy the option")
	}
	cmd := os.Args[1]
	if cmd == "kafka" {
		if CONFIG.Kafka.Topic == "" {
			fmt.Println("Please set the topic")
		}
		if len(CONFIG.Kafka.Brokers) == 0 {
			fmt.Println("Please set the brokers")
		}
		fmt.Printf("kafka topic:%s\n", CONFIG.Kafka.Topic)
		ReadKafka(CONFIG.Kafka.Topic, CONFIG.Kafka.Brokers, CONFIG.Kafka.GroupId)
	} else if cmd == "web" {
		SetDownloadService()
	}
}
