package main

var CONFIG Config
var REFERER = "Referer"
var UA = "PABank Application Security Team"

const ENCRYPT_KEY = "requests2019"
const GET_METHOD = "GET"
const POST_METHOD = "POST"
const DENY_WORDS = "Access denied"

var BLACK_EXTENSIONS = [...]string{".jpg", ".png"}

var NETWORKS [5]string
