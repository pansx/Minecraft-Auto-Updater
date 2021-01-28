package packer

import (
	"encoding/json"
	"fmt"
	"github.com/jlaffaye/ftp"
	"log"
	"net/pansx/updateInfo"
	"net/pansx/utils"
	"os"
	"path"
	"time"
)

type FtpInfo struct {
	Host        string `json:"host"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	TempPath    string `json:"tempPath"`
	PackagePath string `json:"package_path"`
	connection  *ftp.ServerConn
}

func New(path string) *FtpInfo {
	fi := &FtpInfo{}
	err := json.Unmarshal([]byte(utils.ReadStringFromFile(path)), fi)
	if err != nil {
		fmt.Errorf("无法读取上传信息")
	}
	fi.connect()
	return fi
}

func (f *FtpInfo) connect() *ftp.ServerConn {
	c, err := ftp.Dial(f.Host+":21", ftp.DialWithTimeout(5*time.Second), ftp.DialWithDisabledEPSV(true))
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login(f.Username, f.Password)
	if err != nil {
		log.Fatal(err)
	}
	_ = c.MakeDir(f.TempPath)
	f.connection = c
	return c
}

func (f *FtpInfo) upload(i *updateInfo.FileInfo) {
	srcFile := path.Join(f.PackagePath, i.Name)
	data, err := os.Open(srcFile)
	stat, _ := data.Stat()
	destPath := path.Join("/", f.TempPath, i.Name)
	fmt.Println("上传:", srcFile, "=>", destPath, stat.Size())
	err = f.connection.Stor(destPath, data)
	if err != nil {
		panic(err)
	}
}

func (f *FtpInfo) renameToDownload() error {
	downloadPath := "/download"
	err := f.connection.Rename(downloadPath, downloadPath+"-"+utils.GenUUID())
	if err == nil {
		err = f.connection.Rename(f.TempPath, downloadPath)
	}
	return err
}
