package utils

import (
	"archive/zip"
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
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

//ReadStringFromFile 从文件中读取string
func ReadStringFromFile(inFile string) string {
	b, err := ioutil.ReadFile(inFile)
	if err != nil {
		return ""
	}
	return string(b)
}

//ReadStringFromURL 从URL中读取string
func ReadStringFromURL(url string) string {
	r, e := http.Get(url)
	if e != nil {
		return ""
	}
	b, e := ioutil.ReadAll(r.Body)
	if e != nil {
		return ""
	}
	return string(b)
}

//GetFileHashList 返回一个文件夹下所有文件的hash，文件的相对路径为key，hash为value
func GetFileHashList(path string) map[string]string {
	nFile := 0
	c := make(chan [2]string) //用于接收完成信号
	filepath.Walk(path, func(p string, i os.FileInfo, e error) error {
		if !i.IsDir() {
			nFile++
			go func() {
				hash := GetHash(p)      //计算hash，这步是异步的
				c <- [2]string{p, hash} //发送键值对
			}()
		}
		return nil
	})
	m := make(map[string]string)
	if nFile == 0 {
		close(c)
	}
	for knv := range c { //等待异步执行完成
		relativePath := knv[0]
		m[filepath.ToSlash(relativePath)] = knv[1]
		nFile--
		if nFile == 0 {
			close(c)
		}
	}
	return m
}

//WriteStringToFile 将String写入文件，覆盖模式
func WriteStringToFile(file, s string) error {
	var e error
	if IsFileOrDirectoryExists(file) {
		e = os.Remove(file)
		if e != nil {
			return e
		}
	}
	f, e := os.Create(file)
	defer f.Close()
	if e != nil {
		return e
	}
	io.WriteString(f, s)
	return nil
}

//Unzip 解压缩文件，相对路径模式
func Unzip(zipFile string, destDir string) error {
	zipReader, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	for _, f := range zipReader.File {
		fpath := filepath.Join(destDir, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
		} else {
			if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
				return err
			}

			inFile, err := f.Open()
			if err != nil {
				return err
			}
			defer inFile.Close()

			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer outFile.Close()

			_, err = io.Copy(outFile, inFile)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
