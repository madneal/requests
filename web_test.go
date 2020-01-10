package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoGet(t *testing.T) {
	request := Request{
		Url:       "https://www.google.com",
		Headers:   nil,
		Method:    "GET",
		Host:      "www.baidu.com",
		AgentId:   "test",
		Timestamp: 0,
		Postdata:  "",
	}
	res := DoGet(request)
	assert.Equal(t, 200, res.StatusCode(), "the status code should be 200")
}

func TestDoPost(t *testing.T) {
	request := Request{
		Url: "http://111.231.70.62:8000/admin/tokens/new/",
		Headers: map[string]string{
			"Cookie":       "MacaronSession=a3cf053cd0b2d457; user=admin",
			"Content-Type": "application/x-www-form-urlencoded",
		},
		Method:    "POST",
		Host:      "111.231.70.62:8000",
		AgentId:   "TEST",
		Timestamp: 0,
		Postdata:  "_csrf=&tokens=1341234&desc=pinga2345234545234523452345n&type=github",
	}
	res := DoPost(request)
	fmt.Println(string(res.Body()))
	assert.Equal(t, 302, res.StatusCode(), "the status code should be 302")
}

func TestMatchIp(t *testing.T) {
	falseIp := "192.168.21.1"
	assert.Equal(t, false, MatchIp(falseIp), "the ip should not match")
	ip1 := "113.98.55.193"
	ip2 := "113.98.240.33"
	ip3 := "183.62.75.65"
	ip4 := "193.62.75.65"
	ip5 := "210.22.18.193"
	assert.Equal(t, true, MatchIp(ip1), "the ip shoud match")
	assert.Equal(t, true, MatchIp(ip2), "the ip shoud match")
	assert.Equal(t, true, MatchIp(ip3), "the ip shoud match")
	assert.Equal(t, true, MatchIp(ip4), "the ip shoud match")
	assert.Equal(t, true, MatchIp(ip5), "the ip shoud match")
}

//func TestMatchUrl(t *testing.T) {
//	postUrl := "http://www.baidu.com/abc/dfgs?adf=1234"
//	getUrl := "http://www.baidu.com/"
//	assert.Equal(t, false, MatchUrl(postUrl, getUrl), "the result shoud be false")
//	postUrl1 := "http://www.baidu.com/abc/def"
//	getUrl1 := "http://www.baidu.com/abc"
//	assert.Equal(t, true, MatchUrl(postUrl1, getUrl1), "the result should be true")
//	postUrl2 := "https://www.baidu.com/abc/def"
//	getUrl2 := "http://www.baidu.com/abc"
//	assert.Equal(t, false, MatchUrl(postUrl2, getUrl2), "the result should be false")
//}

func TestGetIp(t *testing.T) {
	host := "www.baidu.com"
	ips := GetIp(host)
	for _, ip := range ips {
		fmt.Println(ip)
	}
}

func TestCreateResourceByRequest(t *testing.T) {
	request := Request{
		Url:       "http://www.baidu.com/abc/def?name=123",
		Headers:   nil,
		Method:    "GET",
		Host:      "www.baidu.com",
		AgentId:   "test",
		Timestamp: 0,
		Postdata:  "",
	}
	resource := CreateResourceByRequest(request)
	assert.Equal(t, "www.baidu.com/abc/def", resource.Url, "the url should be the same")
	assert.Equal(t, "www.baidu.com/abc", resource.Firstpath, "the fisrtpath should be the same")
	assert.Equal(t, "GET", resource.Method, "the method should be the same")
}

func TestIsValidReferer(t *testing.T) {
	request := Request{
		Url: "https://www.baidu.com/abc",
		Headers: map[string]string{
			"Referer": "https://www.baidu.com/abc",
		},
		Method:    "GET",
		Host:      "www.baidu.com",
		AgentId:   "test",
		Timestamp: 0,
		Postdata:  "",
	}
	assert.Equal(t, true, IsValidReferer(request), "it should be valid Referer")
	request1 := Request{
		Url: "https://www.baidu.com/abc",
		Headers: map[string]string{
			"Referer": "https://www.baidu.com/",
		},
		Method:    "GET",
		Host:      "www.baidu.com",
		AgentId:   "test",
		Timestamp: 0,
		Postdata:  "",
	}
	assert.Equal(t, false, IsValidReferer(request1), "it should not be valid Referer")
	request2 := Request{
		Url:       "https://www.baidu.com/abc",
		Headers:   nil,
		Method:    "GET",
		Host:      "www.baidu.com",
		AgentId:   "test",
		Timestamp: 0,
		Postdata:  "",
	}
	assert.Equal(t, false, IsValidReferer(request2), "it should not be valid Referer")
}

func TestGetIpFromHost(t *testing.T) {
	host := "192.168.192.1:8080"
	assert.Equal(t, "192.168.192.1", GetIpFromHost(host), "the host should be the same")
	host1 := "192.168.1.1"
	assert.Equal(t, "192.168.1.1", GetIpFromHost(host1), "the host should be the same")
}
