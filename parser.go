package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

// Parser reads a configuration file, parses it and returns the content as an init ServiceConfig struct
type Parser interface {
	Parse(configFile string) (RuleConfig, error)
}

// ParserFunc type is an adapter to allow the use of ordinary functions as subscribers.
// If f is a function with the appropriate signature, ParserFunc(f) is a Parser that calls f.
type ParserFunc func(string) (RuleConfig, error)

// Parse implements the Parser interface
func (f ParserFunc) Parse(configFile string) (RuleConfig, error) { return f(configFile) }

// NewParser creates a new parser using the json library
func NewParser() Parser {
	return NewParserWithFileReader(ioutil.ReadFile)
}

// NewParserWithFileReader returns a Parser with the injected FileReaderFunc function
func NewParserWithFileReader(f FileReaderFunc) Parser {
	return parser{fileReader: f}
}

type parser struct {
	fileReader FileReaderFunc
}

// Parser implements the Parse interface
func (p parser) Parse(configFile string) (RuleConfig, error) {
	var result RuleConfig
	data, err := p.fileReader(configFile)
	if err != nil {
		return result, err
	}
	err = yaml.Unmarshal(data, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// FileReaderFunc is a function used to read the content of a config file
type FileReaderFunc func(string) ([]byte, error)

func parseDuration(v string) time.Duration {
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0
	}
	return d
}
