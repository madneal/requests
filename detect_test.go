package main

import (
	"testing"
)

func TestNewWeakPasswordPlugin(t *testing.T) {
	plugin := NewWeakPasswordPlugin()
	request := Request{
		Postdata: "password=cldNz4uQghdfdsfksdf==",
	}
	isVuln, result := plugin.check(&request)
	Log.Info(isVuln)
	Log.Info(result)
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
	Log.Info(result)
}
