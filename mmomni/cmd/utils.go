package cmd

import (
	"math/rand"
	"strings"
	"time"
)

const PasswdSize = 40

var PasswdRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

func init() {
	rand.Seed(time.Now().UnixNano())
}

func CreatePGPassword() string {
	b := make([]rune, PasswdSize)
	for i := range b {
		b[i] = PasswdRunes[rand.Intn(len(PasswdRunes))]
	}
	return string(b)
}

func ParseFQDN(fqdn string) string {
	fqdn = strings.TrimPrefix(fqdn, "http://")
	fqdn = strings.TrimPrefix(fqdn, "https://")
	return strings.Split(fqdn, "/")[0]
}
