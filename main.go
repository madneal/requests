package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	//"gopkg.in/yaml.v2"
	//"gopkg.in/yaml.v2"
)

func init() {
	source, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Println(err)
	}

	err = yaml.Unmarshal(source, &CONFIG)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(CONFIG.Kafka)
}

func main() {
	if CONFIG.Kafka.Topic == "" {
		fmt.Println("Please set the topic")
	}
	if len(CONFIG.Kafka.Brokers) == 0 {
		fmt.Println("Please set the brokers")
	}
	ReadKafka(CONFIG.Kafka.Topic, CONFIG.Kafka.Brokers)
}
