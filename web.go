package main

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)
import "github.com/go-resty/resty/v2"

type Request struct {
	Url       string
	Headers   map[string]string
	Method    string
	Host      string
	AgentId   string
	Timestamp int64
	Postdata  string
}

func SendRequest(request Request) {
	if request.Method == POST_METHOD {
		results := *MatchUrl(request.Url)
		if len(results) > 0 {
			u, err := url.Parse(request.Url)
			if err != nil {
				fmt.Println(err)
			}
			for _, result := range results {
				resource := Resource{
					Url:       u.Host + u.Path,
					Protocol:  result.Protocol,
					Method:    POST_METHOD,
					Firstpath: "/" + strings.Split(u.Path, "/")[1],
				}
				NewResouce(resource)
			}
		} else {
			return
		}
	}
	if IsNeedReplay(request.Host) == false {
		return
	}
	var res *resty.Response
	if request.Method == GET_METHOD {
		res = DoGet(request)
	} else if request.Method == POST_METHOD {
		//res = DoPost(request)
		fmt.Println("there should not exist any post request")
	} else {
		fmt.Print("method does not support")
	}
	statusCode := res.StatusCode()
	if statusCode == 200 {

	} else {
		return
	}
}

func DoGet(request Request) *resty.Response {
	client := resty.New()
	res := client.R()
	res.SetHeaders(request.Headers)
	response, err := res.Get(request.Url)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response.StatusCode())
	return response
}

func DoPost(request Request) *resty.Response {
	client := resty.New()

	res, err := client.R().SetHeaders(request.Headers).SetBody(request.Postdata).Post(request.Url)
	if err != nil {
		fmt.Println(err)
	}
	return res
}

func CreateResourceByRequest(request Request) Resource {
	u, err := url.Parse(request.Url)
	if err != nil {
		fmt.Println(nil)
		return Resource{}
	}
	path := "/" + strings.Split(u.Path, "/")[1]
	return Resource{
		Url:       u.Host + u.Path,
		Protocol:  u.Scheme,
		Method:    request.Method,
		Firstpath: u.Host + path,
	}
}

func GetIp(host string) []net.IP {
	ip, err := net.LookupIP(host)
	if err != nil {
		fmt.Println(err)
		return []net.IP{}
	}
	return ip
}

// check if ip in the given networks
func MatchIp(ip string) (result bool) {
	result = false
	if len(CONFIG.Network.Network) == 0 {
		fmt.Println("Plase assign network in config.yaml!")
	}
	for _, network := range CONFIG.Network.Network {
		_, subnet, err := net.ParseCIDR(network)
		if err != nil {
			fmt.Println(err)
		}
		if subnet.Contains(net.ParseIP(ip)) {
			result = true
			break
		}
	}
	return result
}

// judge if url need to replay
func IsNeedReplay(host string) bool {
	// judge the ip of host if matches network
	isIp, err := regexp.MatchString("^[0-9]+\\.", host)
	if err != nil {
		fmt.Println(err)
	}
	if isIp == true {
		return true
	} else {
		ips := GetIp(host)
		for ip := range ips {
			ipStr := string(ip)
			if MatchIp(ipStr) == true {
				return true
			}
		}
	}
	return false
}

// judge if urls match, host + only one path
//func MatchUrl(postUrl, getUrl string) bool {
//	uGet, err := url.Parse(getUrl)
//	if err != nil {
//		fmt.Println(err)
//	}
//	uPost, err := url.Parse(postUrl)
//	if err != nil {
//		fmt.Println(err)
//	}
//	if uGet.Path == "" || uPost.Path == "" {
//		return false
//	}
//	pathGet := "/" + strings.Split(uGet.Path, "/")[1]
//	pathPost := "/" + strings.Split(uPost.Path, "/")[1]
//	return (uGet.Host + pathGet) == (uPost.Host + pathPost)
//}
