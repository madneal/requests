package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/segmentio/kafka-go"
	"os"
	"regexp"
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
	//fmt.Printf("process msg: %s\n", msg)
	ParseJson(msg)
}

func ParseJson(msg string) {
	var request Request
	var data map[string]interface{}
	var err error
	if err = json.Unmarshal([]byte(msg), &data); err != nil {
		Log.Error(err)
		return
	}

	if data["agentId"] != nil {
		request.AgentId = data["agentId"].(string)
	}
	if data["t"] != nil {
		request.Timestamp = int64(data["t"].(float64))
	}
	if data["method"] != nil {
		request.Method = data["method"].(string)
	} else {
		return
	}
	//headers := make(map[string]string)
	if data["agentId"] == nil {
		if data["host"] != nil {
			request.Host = data["host"].(string)
		}
		request.Url = ObtainUrl(data)
	}

	if request.Method == "POST" {
		if data["postdata"] != nil {
			request.Postdata = data["postdata"].(string)
		} else {
			return
		}
		pass, result := CheckWeakPass(request.Postdata)
		if result {
			CreateCred(&request, pass)
		}
	}
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

func CheckWeakPass(data string) (string, bool) {
	var pass string
	re := regexp.MustCompile(`(?i)p(ass)?(word|wd)?"?\s?(=|:)+\s?("|')?([0-9a-zA-Z]{1,10})`)
	result := re.FindStringSubmatch(data)
	if len(result) == 0 {
		return pass, false
	}
	return result[len(result)-1], true
}

func CreateCred(request *Request, pass string) {
	cred := Cred{
		Url:         request.Url,
		Password:    pass,
		Postdata:    request.Postdata,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	err := NewCred(&cred)
	if err != nil {
		Log.Error(err)
	}
}
