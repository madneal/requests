package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestSetDownloadService(t *testing.T) {
	SetupServices()
}

func TestIsTokenValid(t *testing.T) {
	assert.Equal(t, true, IsTokenValid("e0b07c477115cd750d494c0dd1b19b03"), "The token should be valid")
}

func TestBatchObtainIp(t *testing.T) {
	assets := make([]Asset, 0)
	assets = append(assets, Asset{
		Host: "www.baidu.com",
	})
	assets = append(assets, Asset{
		Host: "www.taobao.com",
	})
	assets = *BatchObtainIp(&assets)
	assert.Equal(t, true, strings.Contains(assets[0].Ip, "180.101.49.12"), "the ip should be included")
	assert.Equal(t, true, strings.Contains(assets[1].Ip, "101.89.125.238"), "the ip shoud be included")
}

func TestAddQuotesForCsv(t *testing.T) {
	data := []string{"a", "b", "c"}
	//AddQuotesForCsv(&data)
	fmt.Println(data)
}
