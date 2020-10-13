package main

import (
	"net"
	"regexp"
	"strings"
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

func ValidateUrl(url string) bool {
	matched, err := regexp.MatchString(`<|>|;|(:\s+)|'`, url)
	if err != nil {
		Log.Error(err)
	}
	return !matched
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
