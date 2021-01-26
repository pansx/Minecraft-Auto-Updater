package utils

import (
	"crypto/sha1"
	"fmt"
	"io"
	"os"
	"strings"
)

//IsFileOrDirectoryExists 造轮子
func IsFileOrDirectoryExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

//GetHash 获取一个文件的sha1
func GetHash(file string) string {
	sha := sha1.New()
	f, _ := os.Open(file)
	defer f.Close()
	io.Copy(sha, f)
	return strings.ToLower(fmt.Sprintf("%X", sha.Sum(nil)))
}
