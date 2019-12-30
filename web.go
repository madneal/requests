package main

import (
	"fmt"
	"net"
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
	var res *resty.Response
	if request.Method == GET_METHOD {
		res = DoGet(request)
	} else if request.Method == POST_METHOD {
		res = DoPost(request)
	} else {
		fmt.Print("method does not support")
	}
	statusCode := res.StatusCode()
	if statusCode == 200 {

	} else {

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
