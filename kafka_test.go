package main

import (
	"testing"
)
import "github.com/stretchr/testify/assert"

func TestReadKafka(t *testing.T) {
	ReadKafka()
}

func TestCheckWeakPass(t *testing.T) {
	pass := "password=134234"
	matchPass, isWeak := CheckWeakPass(pass)
	assert.Equal(t, "134234", matchPass, "the weak password should be the same")
	assert.Equal(t, true, isWeak, "the weak password should be the same")
	pass1 := "passwd:\"adf234"
	matchPass1, isWeak1 := CheckWeakPass(pass1)
	assert.Equal(t, "adf234", matchPass1, "the weak password should be the same")
	assert.Equal(t, true, isWeak1, "the weak password should be the same")
	pass2 := "pwd: \"aaaa\""
	matchPass2, isWeak2 := CheckWeakPass(pass2)
	assert.Equal(t, "aaaa", matchPass2, "the weak password should be the same")
	assert.Equal(t, true, isWeak2, "the weak password should be the same")
	pass3 := "PWD : \"bbbb\""
	matchPass3, isWeak3 := CheckWeakPass(pass3)
	assert.Equal(t, "bbbb", matchPass3, "the weak password should be the same")
	assert.Equal(t, true, isWeak3, "the weak password should be the same")
}
