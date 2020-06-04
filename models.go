package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/jinzhu/gorm"
	"io"
	"time"
)
import _ "github.com/jinzhu/gorm/dialects/mysql"

type Cred struct {
	Id         int64    `gorm:"type:bigint(20) auto_increment;column:id;primary_key"`
	Url        string   `gorm:"type:varchar(1000);column:url"`
	Password   string    `gorm:"type:varchar(100);column:password"`
	Postdata   string    `gorm:"type:varchar(1000);column:postdata"`
	CreatedTime time.Time `gorm:"created"`
	UpdatedTime time.Time `gorm:"updated"`
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
	//fmt.Println(conStr)
	var err error
	db, err = gorm.Open("mysql", conStr)
	db.DB().SetConnMaxLifetime(time.Minute * 5)
	db.DB().SetMaxIdleConns(5)
	db.DB().SetMaxOpenConns(5)
	if err != nil {
		Log.Error(err)
	}

	if !db.HasTable(&Cred{}) {
		db.CreateTable(&Cred{})
	} else {
		db.AutoMigrate(&Cred{})
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

func NewCred(cred *Cred) error {
	if !Exists(cred.Url) {
		return db.Create(&cred).Error
	} else {
		Log.Warnf("Url %s exists!", cred.Url)
		return nil
	}
}

func Exists(url string) bool {
	var cred Cred
	return !db.Where("url = ?", url).First(&cred).RecordNotFound()
}
