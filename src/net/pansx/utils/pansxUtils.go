package utils

import (
	"crypto/rand"
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

func GenUUID() string {

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
