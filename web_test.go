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
