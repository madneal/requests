package main

var CONFIG Config
var REFERER = "Referer"
var UA = "PABank Application Security Team"
var zeekMsg = [...]string{"Content-Type", "Accept-Encoding", "Referer", "Cookie", "Origin", "Host", "Accept-Language",
	"Accept", "Accept-Charset", "Connection", "User-Agent"}

const ENCRYPT_KEY = "requests2019"
const GET_METHOD = "GET"
const POST_METHOD = "POST"
const DENY_WORDS = "Access denied"
const HEADER_TOKEN = "tkzeek"
const HTTP_SCHEMA = "http://"
const CSV_CONTENT_TYPE = "text/csv"
const JSON_CONTENT_TYPE = "application/json"
const LOG_FILE = "pvs.log"
const TIME_FORMAT = "2006-01-02 15:04:05"
const PRD_ENV = "PRD"
const QA_ENV = "QA"

var BLACK_EXTENSIONS = [...]string{".jpg", ".png"}

var NETWORKS [5]string
