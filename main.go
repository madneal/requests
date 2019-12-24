package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	//"gopkg.in/yaml.v2"
	//"gopkg.in/yaml.v2"
)

type Request struct {
	Url     string
	Headers map[string]string
	Method  string
}

func init() {
	source, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		fmt.Println(err)
	}
	var config Config
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(config.Kafka)
}

func main() {
	fmt.Println(1134)
}
