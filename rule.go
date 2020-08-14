package main

type RuleConfig struct {
	Name     string
	Rule struct{
		Method    string
		Expression string
	}
}
