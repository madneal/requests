package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
		Url:         "www.fff.com",
		Protocol:    "",
		Method:      "",
		Firstpath:   "",
		Ip:          "1.1.1.1",
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
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

func TestCheckResourceOutofdate(t *testing.T) {
	startTime := time.Date(2020, 3, 8, 1, 0, 0, 0, time.UTC)
	endTime := time.Date(2020, 3, 20, 8, 0, 0, 0, time.UTC)
	shouldLargeTenDays := CheckResourceOutofdate(float64(10*24), startTime, endTime)
	assert.Equal(t, true, shouldLargeTenDays, "It should larger than 10 days")
	startTime1 := time.Date(2020, 3, 20, 0, 0, 0, 0, time.UTC)
	endTime1 := time.Date(2020, 3, 21, 0, 0, 0, 0, time.UTC)
	shouldLessTenDays := CheckResourceOutofdate(float64(10*24), startTime1, endTime1)
	assert.Equal(t, false, shouldLessTenDays, "It should less than 10 days")
}
