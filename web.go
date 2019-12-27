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
		doGet(request)
	} else if request.Method == POST_METHOD {
		doPost(request)
	} else {
		fmt.Print("method does not support")
	}
}

func doGet(request Request) {
	client := resty.New()
	res := client.R()
	res.SetHeaders(request.Headers)
	response, err := res.Get(request.Url)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(response.StatusCode())
}

func doPost(request Request) {

}
