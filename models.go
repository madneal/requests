package main

import (
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/jinzhu/gorm"
	"math"
	"net/url"
	"strings"
	"time"
)
import _ "github.com/jinzhu/gorm/dialects/mysql"

type Asset struct {
	Id     int64  `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Url    string `gorm:"type:varchar(1000);column:url"`
	Method string `gorm:"type:varchar(10);column:method"`
	Md5    string `gorm:"type:varchar(100);column:md5"`
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

var db *gorm.DB

func init() {
	conStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", CONFIG.Database.User, CONFIG.Database.Pass,
		CONFIG.Database.Host, CONFIG.Database.Port, CONFIG.Database.Name)
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

func NewAsset(asset Asset) error {
	if !Exists(asset.Md5) {
		return db.Create(&asset).Error
	} else {
		fmt.Println("asset exists!")
		return nil
	}
}

// check if record exists by md5
func Exists(md5 string) bool {
	var asset Asset
	return !db.Where("md5 = ?", md5).First(&asset).RecordNotFound()
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

func getLastUpdatedTime(resource Resource) time.Time {
	err := db.First(&resource).Error
	if err != nil {
		Log.Error(err)
	}
	return resource.UpdatedTime
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
