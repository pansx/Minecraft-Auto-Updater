package packer

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/pansx/fileInfo"
	"net/pansx/updateInfo"
	"net/pansx/utils"
	"os"
	"path"
	"path/filepath"
	"testing"
)

const resourceURL = "http://mcupd.sorazone.com/"
const packagePath = "package"
const updateInfoJsonFile = "update_info.json"

func Pack(init bool) {
	exists := utils.IsFileOrDirectoryExists("game")
	if !exists {
		fmt.Println("当前目录看上去并不是打包器可工作的目录,请将打包器放到game文件夹旁")
		return
	}
	fmt.Println("开始根据当前目录下的game文件夹制作更新包")
	os.RemoveAll(filepath.Join(packagePath, "download"))
	os.MkdirAll(filepath.Join(packagePath, "download"), os.ModePerm)
	fileList := utils.GetFileHashList("game")
	fmt.Println("已获得文件列表")
	if init {
		initUpdateInfo := updateInfo.UpdateInfo{
			GameVersion:       0,
			Mirror:            resourceURL,
			IgnoreList:        []string{},
			PackageIgnoreList: []string{},
			FileInfoList:      []*fileInfo.FileInfo{},
		}
		initUpdateInfo.LoadFileInfoByMap(fileList)
		initUpdInfoJSON, _ := json.Marshal(initUpdateInfo)
		utils.WriteStringToFile(filepath.Join(packagePath, updateInfoJsonFile), string(initUpdInfoJSON))
		println("初始化已完成，请将完善后的update_info.json和file_list.json上传到文件服务器，然后再进行打包")
		return
	}
	var newUpdateInfo updateInfo.UpdateInfo
	err := newUpdateInfo.LoadFromJSON(utils.ReadStringFromFile(filepath.Join(packagePath, updateInfoJsonFile)))
	url := resourceURL + updateInfoJsonFile
	if err != nil {
		fmt.Println("读取本地信息也失败,使用默认地址")
	} else {
		url = newUpdateInfo.Mirror + updateInfoJsonFile
		fmt.Println("使用镜像:", url)
	}

	err = newUpdateInfo.LoadFromJSON(utils.ReadStringFromURL(url))
	if err != nil {
		fmt.Println("读取更新信息时出错，是不是忘了初始化文件服务器？输入--init以获取初始化所需的文件，输入--help获取更多帮助")
		fmt.Println("仅读取本地信息")
	}
	newUpdateInfo.GameVersion++ //自动修改游戏版本
	_ = os.Remove(filepath.Join(packagePath, updateInfoJsonFile))
	newUpdateInfoJSON, _ := json.Marshal(newUpdateInfo)
	utils.WriteStringToFile(filepath.Join(packagePath, updateInfoJsonFile), string(newUpdateInfoJSON)) //写入修改了游戏版本的更新信息文件
	nFiles := len(newUpdateInfo.FileInfoList)
	nFilesTotal := nFiles
	c := make(chan string)
	limitor := make(chan int)
	for _, info := range newUpdateInfo.FileInfoList {
		go func(path, name string, limitor chan int) {
			<-limitor
			destZip := filepath.Join(packagePath, "download", name)
			for utils.IsFileOrDirectoryExists(destZip) { //解决同名文件冲突
				fmt.Printf("遇到冲突，尝试解决中...%s\n", destZip)
				destZip += ".conflict"
			}
			Zip(path, destZip)
			c <- path
		}(info.Path, info.Name, limitor)
	}

	limitor <- 0
	for p := range c {
		nFiles--
		fmt.Fprintf(os.Stdout, "全量压缩中[%v/%v]:%v\n", nFilesTotal-nFiles, nFilesTotal, p)
		if nFiles == 0 {
			close(c)
			fmt.Println("")
		} else {
			limitor <- 0
		}
	}
	fmt.Println("制包完毕，请上传更新后的文件\n你可以选择删除原来的包后上传全量包以节省空间，也可以上传增量包以节省上传时间")
}

//IgnoreFileInFileList 用于排除不需要检测更新的文件或文件夹
func IgnoreFileInFileList(ignoreList *[]string, fileLists []*map[string]string, norm bool) {
	var normIgnoreList []string
	for i := range *ignoreList {
		//(*ignoreList)[i] = filepath.FromSlash((*ignoreList)[i])
		if norm {
			normIgnoreList = append(normIgnoreList, filepath.FromSlash((*ignoreList)[i]))
		} else {
			normIgnoreList = append(normIgnoreList, (*ignoreList)[i])
		}
	} //规范化
	for _, fileList := range fileLists {
		del := make([]string, len(normIgnoreList))
		for kf := range *fileList {
			for _, ki := range normIgnoreList {
				kfn := []rune(kf)
				kin := []rune(ki)
				if len(kin) <= len(kfn) {
					if ki == string(kfn[:len(kin)]) {
						del = append(del, kf)
					}
				}
			}
		}
		for _, d := range del {
			if d != "" {
				delete(*fileList, d)
			}
		}
	}
}

func UploadAll() {
	ftpInfo := New("ftp.json")
	ui := updateInfo.New(path.Join(packagePath, updateInfoJsonFile))
	fi := ui.FileInfoList
	for i, fileInfo := range fi {
		fmt.Println("进度:", i, "/", len(fi), int(math.Floor(float64(i*100)/float64(len(fi)))), "%")
		ftpInfo.upload(fileInfo)
	}
	err := ftpInfo.renameToDownload()
	if err != nil {
		panic(err)
	}
	fmt.Println("进度:", len(fi), "/", len(fi), 100, "%")
}

func TestUpload(t *testing.T) {
	UploadAll()
}

//Zip 用于压缩文件srcFile可以是单文件也可以是目录
func Zip(srcFile string, destZip string) error {
	zipfile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	if err != nil {
		return err
	}
	defer archive.Close()

	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		if info.IsDir() {
			//header.Name += sig
			header.Name = filepath.FromSlash(filepath.ToSlash(header.Name + "/"))
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
		}
		return err
	})

	return err
}
