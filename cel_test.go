package main

import (
	"testing"
)

func TestCheck(t *testing.T) {
	def := make([]InterpretableDefinition, 0)
	def = append(def, InterpretableDefinition{
		CheckExpression: "req_postdata.contains(\"GET\")",
	})
	r := Request{
		Method:   GET_METHOD,
		Postdata: "GET 1341234DFSADFASDFASDF",
	}
	result, err := Check(def, &r)
	Log.Error(err)
	Log.Info(result)
}
