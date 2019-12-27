package main

import (
	"fmt"
	"github.com/jinzhu/gorm"
)
import _ "github.com/jinzhu/gorm/dialects/mysql"

type Asset struct {
	Id     int64  `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Url    string `gorm:"type:varchar(100);column:url`
	Method string `gorm:"type:varchar(5);column:method"`
	Md5    string `gorm:"type:varchar(100);column:md5"`
}

var db *gorm.DB

func init() {
	conStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", CONFIG.Database.User, CONFIG.Database.Pass,
		CONFIG.Database.Host, CONFIG.Database.Port, CONFIG.Database.Name)
	fmt.Println(conStr)
	var err error
	db, err = gorm.Open("mysql", conStr)
	if err != nil {
		fmt.Println(err)
	}
	if !db.HasTable(&Asset{}) {
		db.CreateTable(&Asset{})
	} else {
		db.AutoMigrate(&Asset{})
	}
	//defer db.Close()
}

func NewAsset(asset Asset) error {
	return db.Create(&asset).Error
}

// check if record exists by md5
func Exists(md5 string) bool {
	var asset Asset
	return !db.Where("md5 = ?", md5).First(&asset).RecordNotFound()
}
