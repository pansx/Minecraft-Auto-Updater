package packer

import (
	"encoding/json"
	"fmt"
	"github.com/jlaffaye/ftp"
	"log"
	"net/pansx/fileInfo"
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

func (f *FtpInfo) UploadByFileInfo(i *fileInfo.FileInfo) {
	srcFile := path.Join(f.PackagePath, i.Name)
	destPath := path.Join("/", f.TempPath, i.Name)
	err := f.Upload(srcFile, destPath)
	if err != nil {
		panic(err)
	}
}

func (f *FtpInfo) Upload(srcFile string, destPath string) error {
	data, err := os.Open(srcFile)
	if err != nil {
		panic(err)
	}
	stat, _ := data.Stat()
	fmt.Println("上传:", srcFile, "=>", destPath, stat.Size())
	err = f.connection.Stor(destPath, data)
	return err
}

func (f *FtpInfo) RenameToDownload() error {
	downloadPath := "/download"
	if f.TempPath == downloadPath {
		fmt.Println("临时文件夹和下载文件夹一致,跳过移动")
		return nil
	}
	err := f.connection.Rename(downloadPath, downloadPath+"-"+utils.GenUUID())
	_ = f.connection.MakeDir(downloadPath)
	err = f.connection.Rename(f.TempPath, downloadPath)
	if err == nil {
		fmt.Println("重命名到正式目录成功!")
		return nil
	}

	list, err := f.connection.List(f.TempPath)
	if err == nil {
		for i, entry := range list {
			name := entry.Name
			from := f.TempPath + "/" + name
			to := downloadPath + "/" + name
			err := f.connection.Delete(to)
			err = f.connection.Rename(from, to)
			fmt.Println("移动文件到正式目录:", from, "=>", to, i, "/", len(list))
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	f.connection.RemoveDir(f.TempPath)
	return err
}
