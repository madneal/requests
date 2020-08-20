package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
)

type Plugin struct {
	Name            string
	Type            string
	Expression      string
	check           func(*Request) (bool, string)
	checkExpression func(*string, *Request) (bool, string)
}

type RuleConfig struct {
	Name string
	Rule struct {
		Method     string
		Expression string
	}
}

func InitialRule(yamlName string) *RuleConfig {
	var rule RuleConfig
	source, err := ioutil.ReadFile(yamlName)
	if err != nil {
		Log.Error(err)
		return nil
	}
	err = yaml.Unmarshal(source, &rule)
	return &rule
}

func CheckExpression(express *string, r *Request) (bool, string) {
	def := make([]InterpretableDefinition, 0)
	def = append(def, InterpretableDefinition{
		CheckExpression: *express,
	})
	result, err := Check(def, r)
	if err != nil {
		Log.Error(err)
	}
	return result, r.Postdata
}

func CheckWeakPassword(request *Request) (bool, string) {
	var pass string
	re := regexp.MustCompile(`(?i)(password|passwd|pass|pwd)(["'])?\s?([:=])\s?("|'|)?([0-9a-zA-Z]{1,8})(["'&])+`)
	result := re.FindStringSubmatch(request.Postdata)
	if len(result) == 0 {
		return false, pass
	}
	detail := result[5]
	return true, detail
}

func NewWeakPasswordPlugin() *Plugin {
	return &Plugin{
		Name:  "Weak password",
		check: CheckWeakPassword,
	}
}

func NewYamlPlugin(filename string) *Plugin {
	rule := InitialRule(filename)
	return &Plugin{
		Name:            rule.Name,
		Type:            "yaml",
		Expression:      rule.Rule.Expression,
		checkExpression: CheckExpression,
	}
}

func InitialYamlPlugins() []*Plugin {
	plugins := make([]*Plugin, 0)
	files, err := ioutil.ReadDir("rules")
	if err != nil {
		Log.Error(err)
		return plugins
	}
	for _, file := range files {
		plugins = append(plugins, NewYamlPlugin(filepath.Join("rules", file.Name())))
	}
	return plugins
}

func CheckVulns(req *Request) {
	plugins := make([]*Plugin, 0)
	WeakPasswordPlugin := NewWeakPasswordPlugin()
	plugins = append(plugins, WeakPasswordPlugin)
	plugins = append(plugins, InitialYamlPlugins()...)
	for _, plugin := range plugins {
		var isVuln bool
		var result string
		if "yaml" == plugin.Type {
			isVuln, result = plugin.checkExpression(&plugin.Expression, req)
		} else {
			isVuln, result = plugin.check(req)
		}
		if isVuln {
			vuln := CreateVuln(plugin.Name, result, req)
			err := NewVuln(vuln)
			if err != nil {
				Log.Error(err)
			}
		}
	}
}

func CreateVuln(name, detail string, req *Request) *Vuln {
	return &Vuln{
		Name:    name,
		Detail:  detail,
		ReqStr:  ConvertReqToStr(req),
		Url:     req.Url,
		RespStr: ObtainRespStr(req),
	}
}

func ConvertReqToStr(req *Request) string {
	var result string
	urlParse, err := url.Parse(req.Url)
	if err != nil {
		Log.Errorf("Parse url: %s failed, %+v", req.Url, err)
		return ""
	}

	result += fmt.Sprintf("%s %s?%s HTTP/1.1\n", req.Method, urlParse.Path, urlParse.RawQuery)
	result += fmt.Sprintf("Host: %s\n", req.Host)
	for key, header := range req.Headers {
		result += fmt.Sprintf("%s: %s\n", key, header)
	}
	if req.Method == POST_METHOD {
		result += "\r\n"
		result += req.Postdata + "\n"
	}
	return result
}

func ObtainRespStr(req *Request) string {
	var result string
	result += "Status Code:" + strconv.Itoa(req.StatusCode)
	return result
}
