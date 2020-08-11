package main

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
)

type Plugin struct {
	Name  string
	check func(*Request) (bool, string)
}

func CheckWeakPassword(request *Request) (bool, string) {
	var pass string
	re := regexp.MustCompile(`(?i)(password|passwd|pass|pwd)("|')?\s?(:|=)\s?("|'|)?([0-9a-zA-Z]{1,8})("|'|&)+`)
	result := re.FindStringSubmatch(request.Postdata)
	if len(result) == 0 {
		return false, pass
	}
	detail := fmt.Sprintf("Weak password found! Weakpassword: %s; Postdata: %s", result[5], request.Postdata)
	return true, detail
}

func NewWeakPasswordPlugin() *Plugin {
	return &Plugin{
		Name:  "Weak password",
		check: CheckWeakPassword,
	}
}

func CheckVulns(req *Request) {
	plugins := make([]*Plugin, 0)
	WeackPasswordPlugin := NewWeakPasswordPlugin()
	plugins = append(plugins, WeackPasswordPlugin)
	for _, plugin := range plugins {
		isVuln, result := plugin.check(req)
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
