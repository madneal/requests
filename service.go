package main

import (
	"encoding/csv"
	"net/http"
	"strconv"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	resources, err := QueryAllServices()
	if err != nil {
		Log.Error(err)
	}
	wr := csv.NewWriter(w)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=resources.csv")
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

func AssetsHandler(w http.ResponseWriter, r *http.Request) {
	assets, err := QueryAllAssets()
	if err != nil {
		Log.Error(err)
	}
	wr := csv.NewWriter(w)
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=resources.csv")
	wr.Write([]string{"id", "url", "method"})
	for i := range *assets {
		asset := (*assets)[i]
		record := []string{strconv.Itoa(int(asset.Id)), asset.Url, asset.Method}
		err := wr.Write(record)
		if err != nil {
			Log.Error(err)
			return
		}
	}
	wr.Flush()
}

func SetDownloadService() {
	http.HandleFunc("/download-resources", DownloadHandler)
	http.HandleFunc("/download-assets", AssetsHandler)
	Log.Info(http.ListenAndServe(":80", nil))
}
