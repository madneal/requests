package main

import (
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
	assert.Equal(t, res.StatusCode(), 200, "the status code should be 200")
}
