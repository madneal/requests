package main

var CONFIG Config
var GET_METHOD = "GET"
var POST_METHOD = "POST"
var REFERER = "Referer"
var UA = "PABank Application Security Team"
var zeekMsg = [...]string{"Content-Type", "Accept-Encoding", "Referer", "Cookie", "Origin", "Host", "Accept-Language",
	"Accept", "Accept-Charset", "Connection", "User-Agent"}

const ENCRYPT_KEY = "requests2019"
const DENY_WORDS = "Access denied"
const HEADER_TOKEN = "tkzeek"

var BLACK_EXTENSIONS = [...]string{".jpg", ".png"}

var NETWORKS [5]string
