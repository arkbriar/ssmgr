package main

import (
	"time"
	"math/rand"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

const passwordLength = 10
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

func RandomPassword() string {
	b := make([]rune, passwordLength)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
