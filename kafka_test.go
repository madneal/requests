package main

import "testing"
import "github.com/stretchr/testify/assert"

func TestReadKafka(t *testing.T) {
	var localhost []string
	topic := "test"
	localhost = append(localhost, "localhost:9092")
	ReadKafka(topic, localhost)
}

func TestParseJson(t *testing.T) {
	data := `{
	"url": "http://testasp.vulnweb.com/showthread.asp?id=0",
		"headers": [{
		"name": "Upgrade-Insecure-Requests",
		"value": "1"
	}, {
		"name": "DNT",
			"value": "1"
	}, {
		"name": "User-Agent",
			"value": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Safari/537.36"
	}, {
		"name": "Accept",
			"value": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3"
	}, {
		"name": "Referer",
			"value": "http://testasp.vulnweb.com/showforum.asp?id=1"
	}, {
		"name": "Accept-Encoding",
			"value": "gzip, deflate"
	}, {
		"name": "Accept-Language",
			"value": "zh-CN,zh;q=0.9"
	}, {
		"name": "Cookie",
			"value": "ASPSESSIONIDSADTTCQT=EHPPEOGBOJHEGNLOJFCPCBEK"
	}],
"host": "testasp.vulnweb.com",
"method": "GET",
"agentId": "b77f3736-8542-4626-a9aa-c0fd41d15b61",
"postdata": "",
"t": 1565079259
}`
	data1 := `
{
    "url": "http://testasp.vulnweb.com/showthread.asp?id=0",
    "headers": {
        "Host": "testasp.vulnweb.com",
        "Connection": "keep-alive",
        "Cache-Control": "max-age=0",
        "DNT": "1",
        "Upgrade-Insecure-Requests": "1",
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3",
        "Accept-Encoding": "gzip, deflate",
        "Accept-Language": "zh-CN,zh;q=0.9",
        "Cookie": "ASPSESSIONIDQQBRQABB=FDKAKDNCHPAKFGGIMLNLFBLB"
    },
    "host": "testasp.vulnweb.com",
    "method": "GET",
    "agentId": "b77f3736-8542-4626-a9aa-c0fd41d15b61",
    "postdata": "",
    "t": 1565079259
}
`
	data2 := `
{
	"Content-Type": "json",
	"Referer": "abc",
	"host": "abc",
	"method": "GET",
	"agentId": "14143",
	"Accept-Encoding": "fdfasdsf", 
	"Referer": "fdasdfasdf", 
	"Cookie": "afasdf", 
	"Origin": "fasdfasdf", 
	"Host": "Accept-Language",
	"Accept": "Accept-Charset", 
	"Connection": "User-Agent",
	"Accept-Language": "afdsfa",
	"Accept-Charset": "dfadf",
	"User-Agent": "afdasdf",
	"uri": "fdasdf",
	"t": 1565079259
}
`
	data3 := `
{
    "url": "http://testasp.vulnweb.com/showthread.asp?id=0",
    "headers": {
        "Host": "testasp.vulnweb.com",
        "Connection": "keep-alive",
        "Cache-Control": "max-age=0",
        "DNT": "1",
        "Upgrade-Insecure-Requests": "1",
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3",
        "Accept-Encoding": "gzip, deflate",
        "Accept-Language": "zh-CN,zh;q=0.9",
        "Cookie": "ASPSESSIONIDQQBRQABB=FDKAKDNCHPAKFGGIMLNLFBLB"
    },
    "host": "testasp.vulnweb.com",
    "method": "POST",
    "agentId": "b77f3736-8542-4626-a9aa-c0fd41d15b61",
    "postdata": "YWJjPTEyMw==",
    "t": 1565079259
}
`
	request, _ := ParseJson(data)
	assert.Equal(t, request.Url, "http://testasp.vulnweb.com/showthread.asp?id=0", "the url should be the same")
	assert.Equal(t, request.Headers["Referer"], "http://testasp.vulnweb.com/showforum.asp?id=1", "the Referer should be the same")
	assert.Equal(t, request.Host, "testasp.vulnweb.com", "the host should be the same")
	assert.Equal(t, request.Timestamp, int64(1565079259), "the t should be the same")
	request1, _ := ParseJson(data1)
	assert.Equal(t, request1.Url, "http://testasp.vulnweb.com/showthread.asp?id=0")
	assert.Equal(t, request1.Headers["DNT"], "1", "the header should be the same")
	assert.Equal(t, request1.Method, "GET", "the method should be the same")
	assert.Equal(t, request1.Timestamp, int64(1565079259), "the t should be the same")
	request2, _ := ParseJson(data2)
	assert.Equal(t, request2.Headers["Content-Type"], "json", "the content-type should be the same")
	assert.Equal(t, request2.Timestamp, int64(1565079259), "the timestamp should be the same")
	request3, _ := ParseJson(data3)
	assert.Equal(t, request3.Postdata, "abc=123", "the post data should be the same")
}

func TestInsertAsset(t *testing.T) {
	data := `
{
    "url": "http://testasp.vulnweb.com/showthread.asp?id=0",
    "headers": {
        "Host": "testasp.vulnweb.com",
        "Connection": "keep-alive",
        "Cache-Control": "max-age=0",
        "DNT": "1",
        "Upgrade-Insecure-Requests": "1",
        "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/77.0.3865.90 Safari/537.36",
        "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3",
        "Accept-Encoding": "gzip, deflate",
        "Accept-Language": "zh-CN,zh;q=0.9",
        "Cookie": "ASPSESSIONIDQQBRQABB=FDKAKDNCHPAKFGGIMLNLFBLB"
    },
    "host": "testasp.vulnweb.com",
    "method": "GET",
    "agentId": "b77f3736-8542-4626-a9aa-c0fd41d15b61",
    "postdata": "",
    "t": 1565079259
}`
	request, _ := ParseJson(data)
	InsertAsset(request)
}
