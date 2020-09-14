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
	"strings"
	"time"
)
import _ "github.com/jinzhu/gorm/dialects/mysql"

type Asset struct {
	Id          int64     `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Host        string    `gorm:"type:varchar(100);column:host"`
	Ip          string    `gorm:"type:varchar(100);column:ip"`
	Env         string    `gorm:"type:varchar(10);column:env"`
	Port        int       `gorm:"type:int;column:port"`
	CreatedTime time.Time `gorm:"created"`
	UpdatedTime time.Time `gorm:"updated"`
}

type BlackDomain struct {
	Id   int64  `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Host string `gorm:"type:varchar(100);column:host"`
}

type Vuln struct {
	Id        int64     `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Name      string    `gorm:"type:varchar(25);column:name"`
	Detail    string    `gorm:"type:varchar(300);column:detail"`
	ReqStr    string    `gorm:"type:varchar(1000);column:req_str"`
	Url       string    `gorm:"type:varchar(250);column:url"`
	RespStr   string    `gorm:"type:varchar(200);column:resp_str"`
	CreatedAt time.Time `gorm:"created"`
	UpdatedAt time.Time `gorm:"updated"`
}

type Host struct {
	Id        int64     `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Domain    string    `gorm:"type:varchar(100);column:domain"`
	Ip        string    `gorm:"type:varchar(20);column:ip"`
	CreatedAt time.Time `gorm:"created"`
	UpdatedAt time.Time `gorm:"updated"`
}

var db *gorm.DB

func init() {
	var userDecrypted string
	var passDecrypted string
	if CONFIG.Run.Encrypt {
		userDecrypted = Decrypt(CONFIG.Database.User, ENCRYPT_KEY)
		passDecrypted = Decrypt(CONFIG.Database.Pass, ENCRYPT_KEY)
	} else {
		userDecrypted = CONFIG.Database.User
		passDecrypted = CONFIG.Database.Pass
	}
	conStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", userDecrypted,
		passDecrypted, CONFIG.Database.Host, CONFIG.Database.Port, CONFIG.Database.Name)
	//Log.Info(conStr)
	var err error
	db, err = gorm.Open("mysql", conStr)
	db.DB().SetConnMaxLifetime(time.Minute * 5)
	db.DB().SetMaxIdleConns(5)
	db.DB().SetMaxOpenConns(5)
	if err != nil {
		Log.Error(err)
	}
	if !db.HasTable(&Asset{}) {
		db.CreateTable(&Asset{})
	} else {
		db.AutoMigrate(&Asset{})
	}
	if !db.HasTable(&BlackDomain{}) {
		db.CreateTable(&BlackDomain{})
	} else {
		db.AutoMigrate(&BlackDomain{})
	}
	if !db.HasTable(&Vuln{}) {
		db.CreateTable(&Vuln{})
	} else {
		db.AutoMigrate(&Vuln{})
	}
	if !db.HasTable(&Host{}) {
		db.CreateTable(&Host{})
		db.Model(&Host{}).AddIndex("host_index", "domain")
	} else {
		db.AutoMigrate(&Host{})
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
			Log.Info(err)
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

func NewVuln(vuln *Vuln) error {
	if !ExistsByMultiFields(vuln, vuln.Url, "url", vuln.Name, "name", vuln.Detail, "detail") {
		return db.Create(&vuln).Error
	} else {
		return nil
	}
}

func NewAsset(asset *Asset) error {
	if !ExistsByHostAndPort(asset.Host, asset.Port) {
		return db.Create(&asset).Error
	} else {
		return nil
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

func Delete(host string) error {
	err := db.Where("host = ?", host).Delete(Asset{}).Error
	return err
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
// "a,b,c" == "c,a,b" is true
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

// check if record exists
func Exists(field, fieldName string) bool {
	//db.LogMode(true)
	var asset Asset
	query := fmt.Sprintf("%s = ?", fieldName)
	return !db.Where(query, field).First(&asset).RecordNotFound()
}

func ExistsByMultiFields(item interface{}, field, fieldName, field1, fieldName1, field2, fieldName2 string) bool {
	db.LogMode(true)
	query := fmt.Sprintf("%s = ? and %s = ? and %s = ?", fieldName, fieldName1, fieldName2)
	return !db.Where(query, field, field1, field2).First(item).RecordNotFound()
}

func IsPortZero(md5 string) bool {
	var asset Asset
	return !db.Where("port = 0 and md5 = ?", md5).First(&asset).RecordNotFound()
}

func ExistsByHostAndPort(host string, port int) bool {
	var asset Asset
	return !db.Where("host = ? and port = ?", host, port).First(&asset).RecordNotFound()
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

func QueryAssetHosts() (*[]string, error) {
	var result []string
	err := db.Model(&Asset{}).Group("host").Pluck("host", &result).Error
	return &result, err
}

func QueryHostAndPort() (*[]Asset, error) {
	var result []Asset
	db.LogMode(true)
	err := db.Debug().Table("assets").Group("host, port").Select("host, port").Scan(&result).Error
	return &result, err
}

func QueryAllVulns() (*[]Vuln, error) {
	var result []Vuln
	err := db.Find(&result).Error
	return &result, err
}

// Batch insert asset, only include host and port
// duplicate by host
func BatchInsertAssets(assets *[]Asset) {
	for _, asset := range *assets {
		if !MatchIp(asset.Ip) {
			continue
		}
		if !ExistsByHostAndPort(asset.Host, asset.Port) && asset.Port != 0 {
			err := db.Create(&asset).Error
			if err != nil {
				Log.Error(err)
			}
		}
	}
}
