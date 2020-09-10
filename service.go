package main

import (
	"crypto/md5"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func AssetsHandler(w http.ResponseWriter, r *http.Request) {
	if !IsTokenValid(r.Header.Get(HEADER_TOKEN)) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	host, isParse := r.URL.Query()["host"]
	var assets *[]Asset
	var err error
	if isParse {
		assets, err = QueryAllAssets(host[0])
	} else {
		assets, err = QueryAllAssets("")
	}
	if err != nil {
		Log.Error(err)
		return
	}
	data, err := json.Marshal(assets)
	if err != nil {
		Log.Error(err)
		return
	}
	w.Header().Set("Content-Type", JSON_CONTENT_TYPE)
	w.Write(data)
}

func DownloadAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := QueryAllAssets("")
	if err != nil {
		Log.Error(err)
	}
	wr := csv.NewWriter(w)
	w.Header().Set("Content-Type", CSV_CONTENT_TYPE)
	filename := getFilename("assets")
	w.Header().Set("Content-Disposition", filename)
	wr.Write([]string{"id", "url", "host", "ip", "method", "params", "created_time", "updated_time"})
	for i := range *assets {
		asset := (*assets)[i]
		record := []string{strconv.Itoa(int(asset.Id)), asset.Host, asset.Ip, asset.CreatedTime.Format(TIME_FORMAT),
			asset.UpdatedTime.Format(TIME_FORMAT)}
		err := wr.Write(record)
		if err != nil {
			Log.Error(err)
			return
		}
	}
	wr.Flush()
}

func AddBlackDomainHandler(w http.ResponseWriter, r *http.Request) {
	if !IsTokenValid(r.Header.Get(HEADER_TOKEN)) {
		http.Error(w, DENY_WORDS, http.StatusForbidden)
		return
	}
	if r.Method == POST_METHOD {
		err := r.ParseForm()
		if err != nil {
			Log.Error(err)
			return
		}
		domain := r.Form.Get("host")
		domains := parseDomains(domain)
		for _, blackDomain := range *domains {
			err = NewDomain(&blackDomain)
			if err != nil {
				Log.Error(err)
				return
			}
		}
		fmt.Fprint(w, "Add host Success")
	} else {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func HostsHandler(w http.ResponseWriter, r *http.Request) {
	//if !IsTokenValid(r.Header.Get(HEADER_TOKEN)) {
	//	http.Error(w, DENY_WORDS, http.StatusForbidden)
	//}
	assets, err := QueryHostAndPort()
	if err != nil {
		Log.Error(err)
		http.Error(w, "Query host failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", CSV_CONTENT_TYPE)
	filename := getFilename("hosts")
	w.Header().Set("Content-Disposition", filename)
	wr := csv.NewWriter(w)
	wr.Write([]string{"host", "port"})
	for _, asset := range *assets {
		record := []string{asset.Host, strconv.Itoa(asset.Port)}
		err := wr.Write(record)
		if err != nil {
			Log.Error(err)
			http.Error(w, "Write to csv failed", http.StatusInternalServerError)
		}
	}
	wr.Flush()
}

func DownloadVulnHanlder(w http.ResponseWriter, r *http.Request) {
	if !IsTokenValid(r.Header.Get(HEADER_TOKEN)) {
		http.Error(w, DENY_WORDS, http.StatusForbidden)
		return
	}
	results, err := QueryAllVulns()
	if err != nil {
		Log.Error(err)
		http.Error(w, "Query cred failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", JSON_CONTENT_TYPE)
	data, err := json.Marshal(results)
	if err != nil {
		Log.Error(err)
		return
	}
	w.Write(data)
}

func PostFileHandler(w http.ResponseWriter, r *http.Request) {
	if !IsTokenValid(r.Header.Get(HEADER_TOKEN)) {
		http.Error(w, DENY_WORDS, http.StatusForbidden)
		return
	}
	file, _, err := r.FormFile("host")
	reader := csv.NewReader(file)
	if err != nil {
		Log.Error(err)
		http.Error(w, "post file failed", http.StatusServiceUnavailable)
	}

	lines, err := reader.ReadAll()
	var assets []Asset
	for i, record := range lines {
		if i == 0 {
			continue
		}
		port, err := strconv.Atoi(record[1])
		if err != nil {
			Log.Error(err)
		}
		assets = append(assets, Asset{
			Host:        record[0],
			Port:        port,
			Env:         PRD_ENV,
			CreatedTime: time.Now(),
			UpdatedTime: time.Now(),
		})
	}
	go HandleHosts(&assets)
}

func BatchUpdateIpHandler(w http.ResponseWriter, r *http.Request) {
	if !IsTokenValid(r.Header.Get(HEADER_TOKEN)) {
		http.Error(w, DENY_WORDS, http.StatusForbidden)
		return
	}
	BatchUpdateIp()
}

func BatchUpdateIp() {
	hosts, err := QueryAssetHosts()
	if err != nil {
		Log.Error(err)
	}
	for _, host := range *hosts {
		ip := GetIpStr(host)
		if !CheckIpStrValid(ip) {
			Delete(host)
			continue
		}
		err = UpdateIp(host, ip)
		if err != nil {
			Log.Error(err)
		}
	}
}

func CheckIpStrValid(ipStr string) bool {
	ips := strings.Split(ipStr, ",")
	for _, ip := range ips {
		if !MatchIp(ip) {
			return false
		}
	}
	return true
}

func HandleHosts(assets *[]Asset) {
	assets = BatchObtainIp(assets)
	BatchInsertAssets(assets)
}

func BatchObtainIp(assets *[]Asset) *[]Asset {
	for index, asset := range *assets {
		host := asset.Host
		ip := GetIpStr(host)
		(*assets)[index].Ip = ip
	}
	fmt.Println("Batch obtain ip finished")
	return assets
}

func getFilename(prefix string) string {
	return fmt.Sprintf("attachment;filename=%s-%s.csv", prefix, time.Now().Format("2006-01-02--15:04:05"))
}

// host string is splitted by ","
func parseDomains(host string) *[]BlackDomain {
	domains := make([]BlackDomain, 0)
	hostArr := strings.Split(host, ",")
	for _, domain := range hostArr {
		blackDomain := BlackDomain{
			Host: domain,
		}
		domains = append(domains, blackDomain)
	}
	return &domains
}

func IsTokenValid(token string) bool {
	dateStr := time.Now().Format("2006-01-02")
	str := dateStr + "-zeekpab"
	md5Sum := md5.Sum([]byte(str))
	fmt.Printf("%x", md5Sum)
	return fmt.Sprintf("%x", md5Sum) == token
}

func SetupServices() {
	http.HandleFunc("/download-assets", DownloadAssets)
	http.HandleFunc("/get-assets", AssetsHandler)
	http.HandleFunc("/new-blackdomain", AddBlackDomainHandler)
	http.HandleFunc("/get-assethosts", HostsHandler)
	http.HandleFunc("/download-vulns", DownloadVulnHanlder)
	http.HandleFunc("/post-hostandport", PostFileHandler)
	http.HandleFunc("/batch-update-ip-temp", BatchUpdateIpHandler)
	port := fmt.Sprintf(":%d", CONFIG.Run.Port)
	Log.Info(http.ListenAndServe(port, nil))
}
