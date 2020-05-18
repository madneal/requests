package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"io"
	"math"
	"net/url"
	"strings"
	"time"
)
import _ "github.com/jinzhu/gorm/dialects/mysql"

type Asset struct {
	Id          int64     `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Url         string    `gorm:"type:varchar(1000);column:url"`
	Method      string    `gorm:"type:varchar(10);column:method"`
	Params      string    `gorm:"type:varchar(1000);column:params"`
	Host        string    `gorm:"type:varchar(100);column:host"`
	Ip          string    `gorm:"type:varchar(100);column:ip"`
	CreatedTime time.Time `gorm:"created"`
	UpdatedTime time.Time `gorm:"updated"`
}

type Resource struct {
	Id          int64     `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Url         string    `gorm:"type:varchar(1000);column:url"`
	Protocol    string    `gorm:"type:varchar(10);column:protocol"`
	Method      string    `gorm:"type:varchar(10);column:method"`
	Firstpath   string    `gorm:"type:varchar(100);column:firstpath"`
	Ip          string    `gorm:"type:varchar(20);column:ip"`
	CreatedTime time.Time `gorm:"created"`
	UpdatedTime time.Time `gorm:"updated"`
}

type BlackDomain struct {
	Id   int64  `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Host string `gorm:"type:varchar(100);column:host"`
}

var db *gorm.DB

func init() {
	var userDecrypted string
	var passDecrypted string
	if CONFIG.Run.Encrypt {
		userDecrypted = Decrypt(CONFIG.Database.User, "requests2019")
		passDecrypted = Decrypt(CONFIG.Database.Pass, "requests2019")
	} else {
		userDecrypted = CONFIG.Database.User
		passDecrypted = CONFIG.Database.Pass
	}
	conStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", userDecrypted,
		passDecrypted, CONFIG.Database.Host, CONFIG.Database.Port, CONFIG.Database.Name)
	//fmt.Println(conStr)
	var err error
	db, err = gorm.Open("mysql", conStr)
	if err != nil {
		Log.Error(err)
	}
	if !db.HasTable(&Asset{}) {
		db.CreateTable(&Asset{})
	} else {
		db.AutoMigrate(&Asset{})
	}
	if !db.HasTable(&Resource{}) {
		db.CreateTable(&Resource{})
	} else {
		db.AutoMigrate(&Resource{})
	}
	if !db.HasTable(&BlackDomain{}) {
		db.CreateTable(&BlackDomain{})
	} else {
		db.AutoMigrate(&BlackDomain{})
	}
	//defer db.Close()
	if CONFIG.Run.Redis == true {
		rdb = redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", CONFIG.Redis.Host, CONFIG.Redis.Port), // use default Addr
			Password: CONFIG.Redis.Password,                                      // no password set
			DB:       CONFIG.Redis.Db,                                            // use default DB
		})
		_, err := rdb.Ping().Result()
		if err != nil {
			fmt.Println(err)
		}
		rdb.Expire(CONFIG.Redis.Set, 24*time.Hour)
	}
}

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func Encrypt(data, passphrase string) string {
	data1 := []byte(data)
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		Log.Error(err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		Log.Error(err)
	}
	ciphertext := gcm.Seal(nonce, nonce, data1, nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

func Decrypt(data, passphrase string) string {
	data1, err := base64.StdEncoding.DecodeString(data)
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		Log.Error(err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		Log.Error(err)
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data1[:nonceSize], data1[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		Log.Error(err)
	}
	return string(plaintext)
}

func NewAsset(asset *Asset) error {
	if !AssetExists(asset.Method, asset.Url) {
		return db.Create(&asset).Error
	} else {
		// if asset out of date, try to update the ip of host
		if CheckIfAssetOutofDate(*asset) {
			isNeedUpdateIp, ip := IsIpNeedUpdate((*asset).Host)
			if isNeedUpdateIp {
				err := UpdateIp((*asset).Host, ip)
				if err != nil {
					Log.Error(err)
				}
			}
		}
		newParams := asset.Params
		oldParams := GetParams(asset)
		asset.UpdatedTime = time.Now()
		asset.Params = UpdateParams(oldParams, newParams)
		return db.Save(&asset).Error
	}
}

func NewDomain(domain *BlackDomain) error {
	if !DomainExists(domain.Host) {
		return db.Create(&domain).Error
	} else {
		Log.Warnf("Domain %s exists!!", domain.Host)
		return nil
	}
}

// UpdateHostIfEmpty is utilized to fix for history data where host is empty
func UpdateHostIfEmpty(asset Asset) error {
	err := db.Where("url = ? and method = ?", asset.Url, asset.Method).First(&asset).Error
	if err != nil {
		return err
	}
	if asset.Host == "" {
		u, err := url.Parse(asset.Url)
		if err != nil {
			return nil
		}
		asset.Host = u.Host
	}
	return db.Save(&asset).Error
}

func UpdateIp(host, ip string) error {
	err := db.Table("assets").Where("host = ?", host).Update(Asset{Ip: ip, UpdatedTime: time.Now()}).Error
	return err
}

func IsIpNeedUpdate(host string) (bool, string) {
	freshIp := GetIpStr(host)
	isNeedUpdate := !(CompareStringArr(QueryIp(host), freshIp))
	if isNeedUpdate {
		return isNeedUpdate, freshIp
	} else {
		return isNeedUpdate, ""
	}
}

// CompareStringArr compares two string consists of ele with ","
// "a,b,c" == "c,a,b"
func CompareStringArr(a, b string) bool {
	if a == "" && b == "" {
		return true
	}
	if a == "" && b != "" {
		return false
	}
	arr := strings.Split(a, ",")
	for _, ele := range arr {
		if !strings.Contains(b, ele) {
			return false
		}
	}
	return true
}

func GetParams(asset *Asset) string {
	err := db.First(&asset).Error
	if err != nil {
		Log.Error(err)
		return ""
	}
	return asset.Params
}

// check if record exists
func Exists(field, fieldName string) bool {
	var asset Asset
	query := fmt.Sprintf("%s = ?", fieldName)
	return !db.Where(query, field).First(&asset).RecordNotFound()
}

// check if record exists
func AssetExists(method, url string) bool {
	var asset Asset
	return !db.Where("method = ? and url = ?", method, url).First(&asset).RecordNotFound()
}

func DomainExists(host string) bool {
	var domain BlackDomain
	return !db.Where("host = ?", host).First(&domain).RecordNotFound()
}

func QueryIp(host string) string {
	var ip string
	var asset Asset
	err := db.Where("host = ?", host).First(&asset).Error
	if err != nil {
		Log.Error(err)
		return ""
	}
	if asset.Ip != "" {
		ip = asset.Ip
	}
	return ip
}

// UpdateParams is utilized to update the params according to the new asset
func UpdateParams(oldParams, newParams string) string {
	f := func(c rune) bool {
		return c == ','
	}
	oldParamsArr := strings.FieldsFunc(oldParams, f)
	newParamsArr := strings.Split(newParams, ",")
	for _, param := range newParamsArr {
		if strings.Contains(oldParams, param) {
			continue
		} else {
			oldParamsArr = append(oldParamsArr, param)
		}
	}
	return strings.Join(oldParamsArr, ",")

}

func DeleteIfExists(resource Resource) error {
	if ResourceExists(resource.Url, resource.Protocol, resource.Method) {
		return db.Delete(resource).Error
	}
	return nil
}

func NewResouce(resource Resource) error {
	if !ResourceExists(resource.Url, resource.Protocol, resource.Method) {
		return db.Create(&resource).Error
	} else {
		if CheckIfResourceOutofdate(resource) {
			ip := resource.Ip
			updated := resource.UpdatedTime
			db.First(&resource)
			resource.Ip = ip
			resource.UpdatedTime = updated
			return db.Save(&resource).Error
		}
		return nil
	}
}

func CheckIfResourceOutofdate(resource Resource) bool {
	lastUpdated := getLastUpdatedTime(resource)
	return CheckIfOutofdate(lastUpdated)
}

func CheckIfAssetOutofDate(asset Asset) bool {
	lastUpdated := getLastUpdatedTimeOfAsset(asset)
	return CheckIfOutofdate(lastUpdated)
}

func getLastUpdatedTime(resource Resource) time.Time {
	err := db.First(&resource).Error
	if err != nil {
		Log.Error(err)
	}
	return resource.UpdatedTime
}

func getLastUpdatedTimeOfAsset(asset Asset) time.Time {
	err := db.First(&asset).Error
	if err != nil {
		Log.Error(err)
	}
	return asset.UpdatedTime
}

// CheckIfOutofdate is utilized to check if the last updated time
// larger than 10 days
func CheckIfOutofdate(lastUpdated time.Time) bool {
	return ComputeDuration(float64(10*24), time.Now(), lastUpdated)
}

// ComputeDuration is utilized to compute the duration between startTime and endTime
// id larger than the hours
func ComputeDuration(hours float64, startTime, endTime time.Time) bool {
	diff := endTime.Sub(startTime).Hours()
	return math.Abs(diff) > hours
}

func ResourceExists(url, protocol, method string) bool {
	var reource Resource
	return !db.Where("url = ? and protocol = ? and method = ?", url, protocol, method).
		First(&reource).RecordNotFound()
}

func QueryAllServices() (*[]Resource, error) {
	resources := make([]Resource, 0)
	err := db.Find(&resources).Error
	return &resources, err
}

func QueryAllAssets(host string) (*[]Asset, error) {
	assets := make([]Asset, 0)
	hosts, err := QueryAllHosts()
	if err != nil {
		return nil, err
	}
	if "" == host {
		if len(*hosts) > 0 {
			err = db.Not("host", *hosts).Find(&assets).Error
		} else {
			err = db.Find(&assets).Error
		}
	} else {
		if len(*hosts) > 0 {
			err = db.Not("host", *hosts).Where("host = ?", host).Find(&assets).Error
		} else {
			err = db.Where("host = ?", host).Find(&assets).Error
		}
	}
	return &assets, err
}

func QueryAllHosts() (*[]string, error) {
	var result []string
	err := db.Model(&BlackDomain{}).Pluck("host", &result).Error
	return &result, err
}

func MatchUrl(postUrl string) *[]Resource {
	resources := make([]Resource, 0)
	uPost, err := url.Parse(postUrl)
	if err != nil {
		Log.Error(err)
		return nil
	}
	if uPost.Path == "" {
		return nil
	}
	pathPost := "/" + strings.Split(uPost.Path, "/")[1]
	firstUrl := uPost.Host + pathPost
	err = db.Where("firstpath = ?", firstUrl).Find(&resources).Error
	if err != nil {
		Log.Error(err)
		return nil
	}
	return &resources
}
