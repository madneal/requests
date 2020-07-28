package main

import (
	"fmt"
	"testing"
)

func TestNewWeakPasswordPlugin(t *testing.T) {
	plugin := NewWeakPasswordPlugin()
	request := Request{
		Postdata: "password=1234",
	}
	isVuln, result := plugin.check(&request)
	fmt.Println(isVuln)
	fmt.Println(result)
}
