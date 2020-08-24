package main

import (
	"github.com/go-resty/resty/v2"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Request struct {
	Url        string
	Headers    map[string]string
	Method     string
	Host       string
	AgentId    string
	Port       int
	Timestamp  int64
	Postdata   string
	StatusCode int
}

func SendRequest(request Request) {
	if !ValidateUrl(request.Url) {
		Log.Warnf("The url is invalid, %s", request.Url)
		return
	}
	if request.Method == POST_METHOD {
		if MatchUrl(request.Url) == nil {
			return
		}
		results := *(MatchUrl(request.Url))
		if len(results) > 0 {
			u, err := url.Parse(request.Url)
			if err != nil {
				Log.Error(err)
			} else {
				for _, result := range results {
					resource := Resource{
						Url:         u.Host + u.Path,
						Protocol:    result.Protocol,
						Method:      POST_METHOD,
						Firstpath:   u.Host + "/" + strings.Split(u.Path, "/")[1],
						Port:        request.Port,
						CreatedTime: time.Now(),
						UpdatedTime: time.Now(),
					}
					err := NewResouce(resource)
					if err != nil {
						Log.Error(err)
					}
				}
			}
		} else {
			return
		}
	}
	// the host is invalid
	if request.Host == "-" {
		return
	}
	isNeedReplay, ip := IsNeedReplay(request.Host)
	if isNeedReplay == false {
		Log.Infof("Request to %s will not replay,host: %s\n", request.Url, request.Host)
		return
	}
	var res *resty.Response
	if request.Method == GET_METHOD {
		res = DoGet(request, ip)
	} else if request.Method == POST_METHOD {
		//res = DoPost(request)
		Log.Info("there should not exist any post request")
	} else {
		Log.Info("method does not support")
	}
	if res == nil {
		return
	}
	statusCode := res.StatusCode()
	resource := CreateResourceByRequest(request, ip)
	if statusCode == 200 {
		Log.Infof("Request to %s successful", request.Url)
		err := NewResouce(*resource)
		if err != nil {
			Log.Error(err)
		}
	} else {
		//err := DeleteIfExists(*resource)
		//if err != nil {
		//	Log.Error(err)
		//}
		return
	}
}

func ValidateUrl(url string) bool {
	matched, err := regexp.MatchString(`<|>|;|(:\s+)|'`, url)
	if err != nil {
		Log.Error(err)
	}
	return !matched
}

func DoGet(request Request, ip string) *resty.Response {
	client := resty.New()
	client.SetProxy(CONFIG.Network.Proxy)
	res := client.R()
	res.SetHeaders(request.Headers)
	response, err := res.Get(request.Url)
	if err != nil {
		Log.Error(err)
		return nil
	}
	Log.Infof("Request to %s: %d\n", request.Url, response.StatusCode())
	return response
}

func DoPost(request Request) *resty.Response {
	client := resty.New()

	res, err := client.R().SetHeaders(request.Headers).SetBody(request.Postdata).Post(request.Url)
	if err != nil {
		Log.Error(err)
	}
	return res
}

// judge if referer valid, the host + firstpath of referer and url is same
// return isValid referer and schema
func IsValidReferer(request Request) (bool, string) {
	for k, v := range request.Headers {
		if k == REFERER {
			if IsCommonUrl(request.Url, v) == true {
				scheme, err := GetScheme(v)
				if err != nil {
					Log.Error(err)
					return false, ""
				}
				return true, scheme
			}
		}
	}
	return false, ""
}

func GetScheme(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}
	return u.Scheme, err
}

func CreateResourceByRequest(request Request, ip string) *Resource {
	u, err := url.Parse(request.Url)
	if err != nil {
		Log.Info(err)
		return nil
	}
	path := "/" + strings.Split(u.Path, "/")[1]
	return &Resource{
		Url:         u.Host + u.Path,
		Protocol:    u.Scheme,
		Method:      request.Method,
		Firstpath:   u.Host + path,
		Ip:          ip,
		Port:        request.Port,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
}

func GetIp(host string) []net.IP {
	ip, err := net.LookupIP(host)
	if err != nil {
		Log.Error(err)
		return []net.IP{}
	}
	return ip
}

func GetIpStr(host string) string {
	isIp, err := regexp.MatchString(`^\d{1,3}\.`, host)
	if err != nil {
		Log.Error(err)
		return ""
	}
	if isIp {
		return host
	}
	ips := GetIp(host)
	var result string
	for _, ip := range ips {
		result += ip.String() + ","
	}
	return strings.TrimRight(result, ",")
}

// check if ip in the given networks
func MatchIp(ip string) (result bool) {
	result = false
	if ip == "" {
		return result
	}
	if len(CONFIG.Network.Network) == 0 {
		Log.Info("Please assign network in config.yaml!")
	}
	for _, network := range CONFIG.Network.Network {
		_, subnet, err := net.ParseCIDR(network)
		if err != nil {
			Log.Error(err)
		}
		if subnet.Contains(net.ParseIP(ip)) {
			result = true
			break
		}
	}
	return result
}

// judge if url need to replay
func IsNeedReplay(host string) (bool, string) {
	ips := *GetIpFromHost(host)
	if len(ips) == 0 {
		Log.Warnf("Cannot obtain ip of host: %s", host)
		return false, ""
	}
	for _, ip := range ips {
		if MatchIp(ip) {
			return true, ip
		}
	}
	return false, ips[0]
}

// GetIpFromHost is utilized to parse IP from host
func GetIpFromHost(host string) *[]string {
	ips := make([]string, 0)
	isIp, err := regexp.MatchString("^[0-9]+\\.", host)
	if err != nil {
		Log.Error(err)
	}
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}
	if isIp {
		ips = append(ips, host)
	} else {
		ipArr := GetIp(host)
		for _, ip := range ipArr {
			ips = append(ips, ip.String())
		}
	}
	return &ips
}

// judge if urls match, host + only one path
func IsCommonUrl(url1, url2 string) bool {
	if strings.Contains(url1, "/") == false || strings.Contains(url2, "/") == false {
		return false
	}
	u1, err := url.Parse(url1)
	if err != nil {
		Log.Error(err)
		return false
	}
	u2, err := url.Parse(url2)
	if err != nil {
		Log.Error(err)
		return false
	}
	if u1.Path == "" || u2.Path == "" {
		return false
	}
	var pathGet string
	var pathPost string
	if len(strings.Split(u1.Path, "/")) > 1 && len(strings.Split(u2.Path, "/")) > 1 {
		pathGet = "/" + strings.Split(u1.Path, "/")[1]
		pathPost = "/" + strings.Split(u2.Path, "/")[1]
	} else {
		return false
	}
	return (u1.Host + pathGet) == (u2.Host + pathPost)
}
