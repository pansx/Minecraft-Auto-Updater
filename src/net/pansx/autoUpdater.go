package main

import (
	"fmt"
	"net/pansx/updateInfo"
	"net/pansx/utils"
	"path"
)

const mainDir = "upd"
const updateInfoFile = "update_info.json"

func main() {
	fmt.Println("======Minecraft自动更新器======")
	tempDir := path.Join(mainDir, "temp")
	downloadDir := path.Join(mainDir, "download")
	requiredDir := []string{mainDir, tempDir, downloadDir}
	utils.MakeDirAll(requiredDir)
	info := updateInfo.New(path.Join(mainDir, updateInfoFile))
	if info.Mirror == "" {
		fmt.Println("不使用镜像")
		info.Mirror = "http://mcupd.sorazone.com/"
	}
	{
		downloaded := updateInfo.New(info.Mirror + updateInfoFile)
		fmt.Println("最新版本:", downloaded.GameVersion, ",本地版本:", info.GameVersion)
		if downloaded.GameVersion != 0 {
			info = downloaded
		}
	}
	fmt.Println("下载开始:", info.GameVersion)
	err := info.CheckAndDownloadAll(downloadDir)
	if err != nil {
		fmt.Println(err)
	}
}
