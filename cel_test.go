package main

import (
	"fmt"
	"testing"
)

func TestCheck(t *testing.T) {
	def := make([]InterpretableDefinition, 0)
	def = append(def, InterpretableDefinition{
		CheckExpression: "req_method.contains(\"GET\")",
	})
	r := Request{
		Method: GET_METHOD,
	}
	result, err := Check(def, &r)
	fmt.Println(err)
	fmt.Println(result)
}
