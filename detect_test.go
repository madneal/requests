package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewWeakPasswordPlugin(t *testing.T) {
	plugin := NewWeakPasswordPlugin()
	request := Request{
		Postdata: "password=123456&",
	}
	isVuln, result := plugin.check(&request)
	assert.True(t, true, isVuln, "This is vulnerable")
	assert.Equal(t, "123456", result, "The password should be 123456")
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

func TestInitialYamlPlugins(t *testing.T) {
	plugins := InitialYamlPlugins()
	r := Request{
		Method:   GET_METHOD,
		Postdata: "password 123456",
	}
	result, err := plugins[0].checkExpression(&(plugins[0].Expression), &r)
	fmt.Println(result)
	fmt.Println(err)
}

func TestCheckVulns(t *testing.T) {
	r := Request{
		Method:   GET_METHOD,
		Postdata: "password 123456",
	}
	CheckVulns(&r)
}
