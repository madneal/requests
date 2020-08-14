package main

import (
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
	Log.Info(err)
	Log.Info(result)
}
