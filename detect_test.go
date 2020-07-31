package main

import (
	"fmt"
	"testing"
)

func TestNewWeakPasswordPlugin(t *testing.T) {
	plugin := NewWeakPasswordPlugin()
	request := Request{
		Postdata: "password=1234",
	}
	isVuln, result := plugin.check(&request)
	fmt.Println(isVuln)
	fmt.Println(result)
}

func TestConvertReqToStr(t *testing.T) {
	req := Request{
		Url:    "https://www.baidu.com/abc/def?name=1341",
		Method: GET_METHOD,
		Headers: map[string]string{
			"Content-Type": "Application/json",
		},
	}
	result := ConvertReqToStr(&req)
	fmt.Println(result)
}
