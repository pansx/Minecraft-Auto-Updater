package main

import (
	"fmt"
	"net/pansx/utils"
	"os"
	"path"
)

const mainDir = "upd"

func main() {
	fmt.Println("======Minecraft自动更新器======")
	tempDir := path.Join(mainDir, "temp")
	downloadDir := path.Join(mainDir, "download")
	requiredDir := []string{mainDir, tempDir, downloadDir}
	for _, rDirName := range requiredDir {
		if !utils.IsFileOrDirectoryExists(rDirName) {
			_ = os.Mkdir(rDirName, os.ModePerm)
		}
	}
}
