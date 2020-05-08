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

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	resources, err := QueryAllServices()
	if err != nil {
		Log.Error(err)
		return
	}
	wr := csv.NewWriter(w)
	w.Header().Set("Content-Type", "text/csv")
	filename := fmt.Sprintf("attachment;filename=resources-%s.csv", time.Now().Format("2006-01-02 15:04:05"))
	w.Header().Set("Content-Disposition", filename)
	wr.Write([]string{"id", "url", "protocol", "method", "firstpath", "ip", "created_time", "updated_time"})
	for i := range *resources {
		resource := (*resources)[i]
		record := []string{strconv.Itoa(int(resource.Id)), resource.Url, resource.Protocol, resource.Method,
			resource.Firstpath, resource.Ip, resource.CreatedTime.Format("2006-01-02 15:04:05"),
			resource.UpdatedTime.Format("2006-01-02 15:04:05")}
		err := wr.Write(record)
		if err != nil {
			Log.Error(err)
			return
		}
	}
	wr.Flush()
}

func ResourcesHandler(w http.ResponseWriter, r *http.Request) {
	if !IsTokenValid(r.Header.Get("tkzeek")) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	resources, err := QueryAllServices()
	if err != nil {
		Log.Error(err)
		return
	}
	data, err := json.Marshal(resources)
	if err != nil {
		Log.Error(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func AssetsHandler(w http.ResponseWriter, r *http.Request) {
	if !IsTokenValid(r.Header.Get("tkzeek")) {
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
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func DownloadAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := QueryAllAssets("")
	if err != nil {
		Log.Error(err)
	}
	wr := csv.NewWriter(w)
	w.Header().Set("Content-Type", "text/csv")
	filename := fmt.Sprintf("attachment;filename=assets-%s.csv", time.Now().Format("2006-01-02 15:04:05"))
	w.Header().Set("Content-Disposition", filename)
	wr.Write([]string{"id", "url", "method", "params", "created_time", "updated_time"})
	for i := range *assets {
		asset := (*assets)[i]
		record := []string{strconv.Itoa(int(asset.Id)), asset.Url, asset.Method, asset.Params, asset.CreatedTime.Format("2006-01-02 15:04:05"),
			asset.UpdatedTime.Format("2006-01-02 15:04:05")}
		err := wr.Write(record)
		if err != nil {
			Log.Error(err)
			return
		}
	}
	wr.Flush()
}

func AddBlackDomainHandler(w http.ResponseWriter, r *http.Request) {
	if !IsTokenValid(r.Header.Get("tkzeek")) {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}
	if r.Method == "POST" {
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

// host string is splited by ,
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

func SetDownloadService() {
	http.HandleFunc("/download-resources", DownloadHandler)
	http.HandleFunc("/download-assets", DownloadAssets)
	http.HandleFunc("/get-resources", ResourcesHandler)
	http.HandleFunc("/get-assets", AssetsHandler)
	http.HandleFunc("/new-blackdomain", AddBlackDomainHandler)
	Log.Info(http.ListenAndServe(":80", nil))
}
