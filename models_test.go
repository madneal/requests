package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestExists(t *testing.T) {
	url := "7c031d9efda97b77a7d6"
	exists := Exists(url, "md5")
	Log.Info(exists)
}

func TestMatchUrl(t *testing.T) {
	postUrl := "http://www.baidu.com/abc"
	Log.Info(*MatchUrl(postUrl))
	assert.Equal(t, true, len(*MatchUrl(postUrl)) == 1, "there shoud mathch one")
	postUrl1 := "http://www.baidu.com"
	assert.Equal(t, false, len(*(MatchUrl(postUrl1))) == 1, "there shoud not mathch one")
}

func TestResourceExists(t *testing.T) {
	resource := Resource{
		Url:         "www.fff.com",
		Protocol:    "",
		Method:      "",
		Firstpath:   "",
		Ip:          "2.1.1.1",
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	assert.Equal(t, true, ResourceExists(resource.Url, resource.Protocol, resource.Method),
		"The resource should exist")
}

func TestNewResouce(t *testing.T) {
	resource := Resource{
		Url:         "www.fff.com",
		Protocol:    "",
		Method:      "",
		Firstpath:   "",
		Ip:          "2.1.1.1",
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	err := NewResouce(resource)
	if err != nil {
		Log.Info(err)
	}
}

func TestQueryAllServices(t *testing.T) {
	resources, err := QueryAllServices()
	if err != nil {
		Log.Info(err)
	}
	Log.Info(len(*resources))
}

func TestQueryAllAssets(t *testing.T) {
	assets, err := QueryAllAssets("www")
	if err != nil {
		Log.Info(err)
	}
	Log.Info(len(*assets))
	fmt.Printf("%v", assets)
}

func TestQueryAllHosts(t *testing.T) {
	domains, err := QueryAllHosts()
	if err != nil {
		Log.Info(err)
	}
	fmt.Printf("%v", domains)
}

func TestCheckResourceOutofdate(t *testing.T) {
	startTime := time.Date(2020, 3, 8, 1, 0, 0, 0, time.UTC)
	endTime := time.Date(2020, 3, 20, 8, 0, 0, 0, time.UTC)
	shouldLargeTenDays := ComputeDuration(float64(10*24), startTime, endTime)
	assert.Equal(t, true, shouldLargeTenDays, "It should larger than 10 days")
	startTime1 := time.Date(2020, 3, 20, 0, 0, 0, 0, time.UTC)
	endTime1 := time.Date(2020, 3, 21, 0, 0, 0, 0, time.UTC)
	shouldLessTenDays := ComputeDuration(float64(10*24), startTime1, endTime1)
	assert.Equal(t, false, shouldLessTenDays, "It should less than 10 days")
}

func TestCheckIfOutofdate(t *testing.T) {
	lastUpdated := time.Date(2020, 3, 25, 0, 0, 0, 0, time.Local)
	result := CheckIfOutofdate(lastUpdated)
	assert.Equal(t, false, result, "The last updated time is less than 10 days")
	lastUpdated1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	result1 := CheckIfOutofdate(lastUpdated1)
	assert.Equal(t, true, result1, "The last updated time is larger than 10 days")
}

func TestCheckIfResourceOutofdate(t *testing.T) {
	testTime := time.Date(2020, 2, 1, 0, 0, 0, 0, time.Local)
	testResource := Resource{
		Url:         "www.test.com",
		Protocol:    "http",
		Method:      GET_METHOD,
		Firstpath:   "aaaa",
		CreatedTime: testTime,
		UpdatedTime: testTime,
	}
	err := NewResouce(testResource)
	if err != nil {
		Log.Error(err)
	}
	result := CheckIfResourceOutofdate(testResource)
	db.Delete(testResource)
	assert.Equal(t, true, result, "The resource should be out of date")

	testTime1 := time.Date(2020, 3, 25, 0, 0, 0, 0, time.Local)
	resource := Resource{
		Url:         "www.test.com.aaa",
		Protocol:    "http",
		Method:      GET_METHOD,
		Firstpath:   "SS",
		CreatedTime: testTime1,
		UpdatedTime: testTime1,
	}
	err1 := NewResouce(resource)
	if err1 != nil {
		Log.Error(err)
	}
	result1 := CheckIfResourceOutofdate(resource)
	db.Delete(resource)
	assert.Equal(t, false, result1, "The resource should not be out of date")
}

func TestDeleteIfExists(t *testing.T) {
	resource := Resource{
		Url:         "www.fff.com",
		Protocol:    "",
		Method:      "",
		Firstpath:   "",
		Ip:          "2.1.1.1",
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	err := DeleteIfExists(resource)
	if err != nil {
		Log.Error(err)
	}
	db.First(&resource)
}

func TestGetIps(t *testing.T) {
	ip := QueryIp("www.baidu.com")
	assert.Equal(t, "1.1.1.1", ip, "the ip shoule be the same")
	ip1 := QueryIp("google.com")
	assert.Equal(t, "", ip1, "the ip should not exist")
}

func TestIsIpNeedUpdate(t *testing.T) {
	host := "www.baidu.com"
	isneedUpdate, _ := IsIpNeedUpdate(host)
	assert.Equal(t, true, isneedUpdate, "The host need to be updated")
	host1 := "taobao.com"
	isNeedUpdate, _ := IsIpNeedUpdate(host1)
	assert.Equal(t, false, isNeedUpdate, "The host doesn't need to be updated")
}

func TestCompareStringArr(t *testing.T) {
	assert.Equal(t, true, CompareStringArr("", ""), "The empty string should be the same")
	assert.Equal(t, true, CompareStringArr("a", "a"), "The two words shoult be the same")
	assert.Equal(t, false, CompareStringArr("a", "b"), "The two words should not be the same")
	assert.Equal(t, true, CompareStringArr("a,b,c", "a,c,b"), "The two string should be the same")
}

func TestUpdateIp(t *testing.T) {
	UpdateIp("www.baidu.com", "2.2.2.2")
}

func TestDecryptPass(t *testing.T) {
	cryped := Encrypt("res", ENCRYPT_KEY)
	Log.Info(cryped)
	decyped := Decrypt(cryped, ENCRYPT_KEY)
	assert.Equal(t, "res", decyped, "the password should be decrypted")
}

func TestQueryAssetHosts(t *testing.T) {
	hosts, _ := QueryAssetHosts()
	Log.Info(*hosts)
}

func TestQueryHostAndPort(t *testing.T) {
	results, _ := QueryHostAndPort()
	Log.Info(*results)
}

func TestBatchInsertAssets(t *testing.T) {
	asset := Asset{
		Host:        "www.google.com",
		Port:        1234,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	asset1 := Asset{
		Host:        "www.micro.com",
		Port:        80,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	assets := make([]Asset, 0)
	assets = append(assets, asset)
	assets = append(assets, asset1)
	BatchInsertAssets(&assets)
	asset2 := Asset{
		Host:        "www.google.com",
		Port:        80,
		CreatedTime: time.Now(),
		UpdatedTime: time.Now(),
	}
	assets = append(assets, asset2)
	BatchInsertAssets(&assets)
	assert.Equal(t, true, ExistsByHostAndPort(asset2.Host, asset2.Port), "The host and port exists")
	for _, asset := range assets {
		db.Where("host = ? and port = ?", asset.Host, asset.Port).Delete(&asset)
	}
}

func TestIsPortZero(t *testing.T) {
	var asset Asset
	result := IsPortZero("7c031d9efda97b77a7d63dd0315e36fa")
	assert.Equal(t, true, result, "the result should be true")
	db.Model(&asset).Where("md5 = ?", "7c031d9efda97b77a7d63dd0315e36fa").Update("port", 1234)
}

func TestCheckPortOfResource(t *testing.T) {
	resource := Resource{
		Url: "www.baidu.com",
	}
	result := CheckPortOfResource(resource)
	assert.Equal(t, true, result, "the result should be true")
}

func TestUpdatePort(t *testing.T) {
	resource := Resource{
		Url:  "www.baidu.com",
		Port: 1234,
	}
	UpdatePort(resource)
}

func TestQueryAllResults(t *testing.T) {
	var result *[]Vuln
	result, err := QueryAllVulns()
	Log.Info(result)
	if err != nil {
		Log.Error(err)
	}
}

func TestDelete(t *testing.T) {
	err := Delete("www.baidu.com")
	if err != nil {
		Log.Error(err)
	}
}
