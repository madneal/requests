package main

import (
	//"fmt"
	"fmt"
	"testing"
)

//func init() {
//	source, err := ioutil.ReadFile("config.yaml")
//	if err != nil {
//		fmt.Println(err)
//	}
//	var config Config
//	err = yaml.Unmarshal(source, &config)
//	if err != nil {
//		fmt.Println(err)
//	}
//	fmt.Println(config.Kafka)
//}

func TestNewAsset(t *testing.T) {
	asset := Asset{
		//Id: 1,
		Url:    "www.baidu.com",
		Method: "GET",
	}
	err := NewAsset(asset)
	if err != nil {
		fmt.Print(err)
	}
	//fmt.Print(result)
}

func TestExists(t *testing.T) {
	md5 := "123"
	exists := Exists(md5)
	fmt.Println(exists)
}
