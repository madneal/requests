package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

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

func TestMatchUrl(t *testing.T) {
	postUrl := "http://www.baidu.com/abc"
	assert.Equal(t, true, len(*MatchUrl(postUrl)) == 1, "there shoud mathch one")
	postUrl1 := "http://www.baidu.com"
	assert.Equal(t, false, len(*MatchUrl(postUrl1)) == 1, "there shoud not mathch one")
}

func TestNewResouce(t *testing.T) {
	resource := Resource{
		Url:       "www.baidu.com",
		Protocol:  "",
		Method:    "",
		Firstpath: "",
		Ip:        "1.1.1.1",
	}
	err := NewResouce(resource)
	if err != nil {
		fmt.Println(err)
	}
}

func TestQueryAllServices(t *testing.T) {
	resources, err := QueryAllServices()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(len(*resources))
}
