package main

import "testing"

func TestParser_Parse(t *testing.T) {
	p := NewParser()
	p.Parse("./rules/demo-rule.yaml")
}
