package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetDownloadService(t *testing.T) {
	SetDownloadService()
}

func TestIsTokenValid(t *testing.T) {
	assert.Equal(t, true, IsTokenValid("e0b07c477115cd750d494c0dd1b19b03"), "The token should be valid")
}
