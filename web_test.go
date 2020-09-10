package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchIp(t *testing.T) {
	ip4 := ""
	assert.Equal(t, false, MatchIp(ip4), "the ip empty should not match")
	ip5 := "1.192.158.172"
	assert.Equal(t, false, MatchIp(ip5), "the ip should not match")
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
		Log.Info(ip)
	}
}

func TestCreateResourceByRequest(t *testing.T) {
	request := Request{
		Url:       "http://www.baidu.com/abc/def?name=123",
		Headers:   nil,
		Method:    GET_METHOD,
		Host:      "www.baidu.com",
		AgentId:   "test",
		Timestamp: 0,
		Postdata:  "",
	}
	resource := CreateResourceByRequest(request, "1.1.1.1")
	assert.Equal(t, "www.baidu.com/abc/def", resource.Url, "the url should be the same")
	assert.Equal(t, "www.baidu.com/abc", resource.Firstpath, "the fisrtpath should be the same")
	assert.Equal(t, GET_METHOD, resource.Method, "the method should be the same")
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
	Log.Info(ipStr)
}

func TestValidateUrl(t *testing.T) {
	url := "http://www.baidu.com/<script>alet"
	result := ValidateUrl(url)
	assert.Equal(t, false, result, "the url should be invalid")
	url1 := "http://www.baidu.com/;;;aaaa"
	assert.Equal(t, false, ValidateUrl(url1), "the url should be invalid")
	url2 := "https://www.baidu.com/Content-lenght: 1234"
	assert.Equal(t, false, ValidateUrl(url2), "the result should be invalid")
}
