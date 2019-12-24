package main

import (
	"fmt"
	"io/ioutil"
)

func ReadFile(filepath string) []byte {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		fmt.Println(err)
	}
	return data
}
