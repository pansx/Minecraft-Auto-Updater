package main

import (
	"fmt"
	"net/pansx/utils"
	"os"
)

func main() {
	fmt.Println("======Minecraft自动更新器======")
	requiredDir := []string{"game", "rubbish", "download"}
	for _, rDirName := range requiredDir {
		if !utils.IsFileOrDirectoryExists(rDirName) {
			_ = os.Mkdir(rDirName, os.ModePerm)
		}
	}
}
