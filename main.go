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
}

func main() {
	Log.Info("*********Begin the Assets detect*************")
	if len(os.Args) < 2 {
		Log.Info("Please speficy the option")
	}
	cmd := os.Args[1]
	if cmd == "kafka" {
		if CONFIG.Kafka.Topic == "" {
			Log.Info("Please set the topic")
		}
		if len(CONFIG.Kafka.Brokers) == 0 {
			Log.Info("Please set the brokers")
		}
		fmt.Printf("kafka topic:%s\n", CONFIG.Kafka.Topic)
		ReadKafka()
	} else if cmd == "web" {
		SetupServices()
	} else if cmd == "e" {
		Log.Info(Encrypt(os.Args[2], ENCRYPT_KEY))
	} else if cmd == "d" {
		Log.Info(Decrypt(os.Args[2], ENCRYPT_KEY))
	}
}
