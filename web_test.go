package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDoGet(t *testing.T) {
	request := Request{
		Url:       "https://www.baidu.com",
		Headers:   nil,
		Method:    "GET",
		Host:      "www.baidu.com",
		AgentId:   "test",
		Timestamp: 0,
		Postdata:  "",
	}
	res := DoGet(request, "1.1.1.1")
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
	ip2 := "113.98.240.35"
	ip3 := "183.62.75.65"
	assert.Equal(t, true, MatchIp(ip1), "the ip shoud match")
	assert.Equal(t, true, MatchIp(ip2), "the ip shoud match")
	assert.Equal(t, true, MatchIp(ip3), "the ip shoud match")
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
	resource := CreateResourceByRequest(request, "1.1.1.1")
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
	result, _ := IsValidReferer(request)
	assert.Equal(t, true, result, "it should be valid Referer")
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
	result1, _ := IsValidReferer(request1)
	assert.Equal(t, false, result1, "it should not be valid Referer")
	request2 := Request{
		Url:       "https://www.baidu.com/abc",
		Headers:   nil,
		Method:    "GET",
		Host:      "www.baidu.com",
		AgentId:   "test",
		Timestamp: 0,
		Postdata:  "",
	}
	result2, _ := IsValidReferer(request2)
	assert.Equal(t, false, result2, "it should not be valid Referer")
}

func TestGetIpFromHost(t *testing.T) {
	host := "192.168.192.1:8080"
	assert.Equal(t, "192.168.192.1", (*GetIpFromHost(host))[0], "the host should be the same")
	host1 := "192.168.1.1"
	assert.Equal(t, "192.168.1.1", (*GetIpFromHost(host1))[0], "the host should be the same")
	host2 := "baidu.com"
	assert.Equal(t, "220.181.38.148", (*GetIpFromHost(host2))[0], "the host should be the same")
}

func TestIsNeedReplay(t *testing.T) {
	host := "192.168.0.1:9090"
	result, ip := IsNeedReplay(host)
	assert.Equal(t, false, result, "the result should be false")
	assert.Equal(t, "192.168.0.1", ip, "the ip should be the same")
	host1 := "113.98.55.192:8080"
	result1, ip1 := IsNeedReplay(host1)
	assert.Equal(t, true, result1, "the result should be true")
	assert.Equal(t, "113.98.55.192", ip1, "the ip should be the same")
}

func TestGetIpStr(t *testing.T) {
	ipStr := GetIpStr("www.baidu.com")
	fmt.Println(ipStr)
}
