package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"testing"
	"time"
)
import "github.com/stretchr/testify/assert"

func TestReadKafka(t *testing.T) {
	ReadKafka()
}

func TestParseJson(t *testing.T) {
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
    "method": GET_METHOD,
    "agentId": "b77f3736-8542-4626-a9aa-c0fd41d15b61",
    "postdata": "",
    "t": 1565079259
}
`
	data2 := `
{
	"Content-Type": "json",
	"Referer": "https://www.baidu.com",
	"host": "www.baidu.com",
	"method": GET_METHOD,
	"agentId": "14143",
	"Accept-Encoding": "utf-8", 
	"Cookie": "name=1324", 
	"Origin": "https://www.baidu.com", 
	"Host": "www.baidu.com",
	"Accept": "*", 
	"Connection": "keep-live",
	"Accept-Language": "ch",
	"Accept-Charset": "ios",
	"User-Agent": "chrome",
	"uri": "/test/test1",
	"resp_p": "80",
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
    "method": POST_METHOD,
    "agentId": "b77f3736-8542-4626-a9aa-c0fd41d15b61",
    "postdata": "YWJjPTEyMw==",
    "t": 1565079259
}
`
	data4 := `
{
	"host": "www.baidu.comhttp:",
	"uri": "/abc/def?name=134",
	"resp_p": "80",
	"method": GET_METHOD
}
`
	data5 := "{\"resp_p\": \"80\"}"
	request1, _ := ParseJson(data1)
	assert.Equal(t, request1.Url, "http://testasp.vulnweb.com/showthread.asp?id=0")
	assert.Equal(t, request1.Headers["DNT"], "1", "the header should be the same")
	assert.Equal(t, request1.Method, GET_METHOD, "the method should be the same")
	assert.Equal(t, request1.Timestamp, int64(1565079259), "the t should be the same")

	request2, _ := ParseJson(data2)
	assert.Equal(t, "json", request2.Headers["Content-Type"], "the content-type should be the same")
	assert.Equal(t, request2.Timestamp, int64(1565079259), "the timestamp should be the same")
	assert.Equal(t, UA, request2.Headers["User-Agent"], "the ua should be the same")

	request3, _ := ParseJson(data3)
	assert.Equal(t, request3.Postdata, "abc=123", "the post data should be the same")

	request4, err := ParseJson(data4)
	fmt.Println(err)
	assert.Equal(t, "http://www.baidu.com:80/abc/def?name=134", request4.Url, "the url should be the same")

	request5, _ := ParseJson(data5)
	fmt.Println(request5)
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
    "method": GET_METHOD,
    "agentId": "b77f3736-8542-4626-a9aa-c0fd41d15b61",
    "postdata": "",
    "t": 1565079259
}`
	request, _ := ParseJson(data)
	InsertAsset(request)
}

func TestRedis(t *testing.T) {
	rdb = redis.NewClient(&redis.Options{
		Addr: "127.0.0.1:6379", // use default Addr
		DB:   0,                // use default DB
	})
	pong, err := rdb.Ping().Result()
	if err != nil {
		Log.Error(err)
	}
	fmt.Println(pong)
	rdb.Expire(CONFIG.Redis.Set, 24*time.Hour)
	url := "wwww.baidu.com"
	err = rdb.SAdd(CONFIG.Redis.Set, url).Err()
	if err != nil {
		fmt.Println(err)
	}
	assert.Equal(t, true, rdb.SIsMember(CONFIG.Redis.Set, url).Val(), "the url should exists")
	assert.Equal(t, false, rdb.SIsMember(CONFIG.Redis.Set, "25542352345").Val(), "the data "+
		"should not exists")
}

func TestSetUrlByScheme(t *testing.T) {
	url, _ := SetUrlByScheme("http", "https://play.golang.org/")
	fmt.Println(url)
	assert.Equal(t, "http://play.golang.org/", url, "the url shoule be http")
}

//func TestCreateAssetByUrl(t *testing.T) {
//	urlStr := ""
//	asset := CreateAssetByUrl(urlStr)
//	assert.Equal(t, "http://gitlab.com/pa/edf/aa", asset.Url, "the url should be the same")
//	assert.Equal(t, "ff,name,bcd", asset.Params, "the params should be the same")
//}

func TestObtainIp(t *testing.T) {
	ip := ObtainIp("wwww.baidu.com")
	assert.Equal(t, "1.1.1.1", ip, "the ip should be the same")
	ip1 := ObtainIp("taobao.com")
	assert.Equal(t, "140.205.94.189", ip1, "the ip should be looked up by dns")
}

func TestRunTask(t *testing.T) {
	msg := "{\"id.resp_p\": 80, \"method\": \"POST\", \"host\": \"bill.sdb.com\", \"postdata\": \"password=123\"}"
	RunTask(msg)
}

func TestCheckIfBlackExtension(t *testing.T) {
	url := "http://wwww.baidu.com"
	result := CheckIfBlackExtension(url)
	assert.Equal(t, false, result, "the url does not contain the black extension")
	url1 := "http://www.baidu.com/1341234.png"
	result1 := CheckIfBlackExtension(url1)
	assert.Equal(t, true, result1, "the url contains the black extension")
}

func TestComputeHash(t *testing.T) {
	hash := ComputeHash("http://www.baidu.comGET")
	fmt.Println(hash)
	assert.True(t, true, 20 == len(hash), "the length should be 20")
}

func TestValidateHost(t *testing.T) {
	host := "wwww.baidu.com"
	assert.Equal(t, true, ValidateHost(host), "the host should be valid")
	host1 := "d12:243:134"
	assert.Equal(t, false, ValidateHost(host1), "the host should be invalid")
	host2 := "192.168.1.1"
	assert.Equal(t, true, ValidateHost(host2), "the host should be valid")
	host3 := "abc"
	assert.Equal(t, false, ValidateHost(host3), "the host should not be valid")
	host4 := "www.baidu.com.net"
	assert.Equal(t, false, ValidateHost(host4), "the host should not be valid")
	host5 := "baidu.comhttp:"
	assert.Equal(t, false, ValidateHost(host5), "the host should not be valid")

	host6 := "219.133.104.74http:"
	assert.Equal(t, false, ValidateHost(host6), "the host should not be valid")

	host7 := "www.pafacebank.com"
	assert.Equal(t, true, ValidateHost(host7), "the host should be valid")

	host8 := "www.baidu.com.cn:80"
	assert.Equal(t, true, ValidateHost(host8), "the host should be valid")

	host9 := "121.90.89.91:8000"
	assert.Equal(t, true, ValidateHost(host9), "the host should be valid")
}
