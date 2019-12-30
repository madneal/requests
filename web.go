package main

import "fmt"
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
	if request.Method == GET_METHOD {
		DoGet(request)
	} else if request.Method == POST_METHOD {
		DoPost(request)
	} else {
		fmt.Print("method does not support")
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
