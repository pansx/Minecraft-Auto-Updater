package main

import (
	"archive/zip"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const resourceURL = "https://minecraft-updater.oss-cn-shanghai.aliyuncs.com/"

//IsFileOrDirectoryExists 造轮子
func IsFileOrDirectoryExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return false
}

//Zip 用于压缩文件srcFile可以是单文件也可以是目录
func Zip(srcFile string, destZip string) error {
	zipfile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	filepath.Walk(srcFile, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		/*
			替换分隔符，现在用自带方法
			sig := "/"
			if runtime.GOOS == "windows" {
				sig = `\`
			}
			header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+sig)
		*/
		header.Name = strings.TrimPrefix(path, filepath.FromSlash(filepath.ToSlash(filepath.Dir(srcFile)+"/")))
		//header.Name = path
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

//GetHash 获取一个文件的sha1
func GetHash(file string) string {
	sha := sha1.New()
	f, _ := os.Open(file)
	defer f.Close()
	io.Copy(sha, f)
	return strings.ToLower(fmt.Sprintf("%X", sha.Sum(nil)))
}

//GetFileList 返回一个文件夹下所有文件的hash，文件的相对路径为key，hash为value
func GetFileList(path string) *map[string]string {
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
		m[knv[0]] = knv[1]
		nFile--
		if nFile == 0 {
			close(c)
		}
	}
	return &m
}

//DownLoadFile 下载文件
func DownLoadFile(url, destDir string) error {
	if IsFileOrDirectoryExists(destDir) {
		return nil
	}
	destFile, err := os.Create(destDir)
	defer destFile.Close()
	if err != nil {
		return err
	}
	var res *http.Response
	res, err = http.Get(url)
	if err != nil {
		return err
	}
	_, err = io.Copy(destFile, res.Body)
	return err
}

//DownLoadFileAndCheck 下载文件并校验hash是否相符
func DownLoadFileAndCheck(url, destDir, hash string) error {
	//log.Println("下载文件并检查 url:" + url + " dest:" + destDir)
	hash = strings.ToLower(hash)
	if IsFileOrDirectoryExists(destDir) {
		if hash == GetHash(destDir) {
			//log.Println("文件校验通过 url:" + url + " dest:" + destDir)
			return nil
		}
		os.Remove(destDir)
		//log.Println("文件校验不通过，重新下载 url:" + url + " dest:" + destDir)
	}
	for i := 0; ; i++ {
		err := DownLoadFile(url, destDir)
		if err == nil {
			if hash == GetHash(destDir) {
				//log.Println("文件校验通过 url:" + url + " dest:" + destDir)
				return nil
			}
		} else {
			if i > 10 { //最多尝试十次
				//log.Println("超过最大重试次数 url:" + url + " dest:" + destDir)
				return err
			}
		}
	}
}

type updateInfo struct {
	GameVersion int      `json:"version"`
	ResourceURL string   `json:"resource_url"`
	IgnoreList  []string `json:"ignore_list"`
}

//返回的是json形式的表达
func (ui updateInfo) String() string {
	d, _ := json.Marshal(ui)
	return string(d)
}

//LoadFromJSON 从json中加载updateInfo
func (ui *updateInfo) LoadFromJSON(s string) error {
	err := json.Unmarshal([]byte(s), ui)
	return err
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

//GetFileListFromJSON 从json中读取filelist
func GetFileListFromJSON(s string) *map[string]string {
	m := make(map[string]string)
	json.Unmarshal([]byte(s), &m)
	return &m
}

//CompareFileList 对比新旧两个文件列表，输出多余文件列表与缺失文件列表
func CompareFileList(localFileList, newFileList *map[string]string) (surp, lack *map[string]string) {
	su := make(map[string]string)
	la := make(map[string]string)
	c := make(chan int)
	go func() {
		for k, vl := range *localFileList {
			vn, exists := (*newFileList)[k]
			if (!exists) || vn != vl { //newFileList里不存在或者hash对不上的文件被认为是多余的
				su[k] = vl
			}
		}
		c <- 0
	}()
	go func() {
		for k, vn := range *newFileList {
			vl, exists := (*localFileList)[k]
			if (!exists) || vn != vl { //localFileList里不存在或者hash对不上的文件被认为是缺失的
				la[k] = vn
			}
		}
		c <- 0
	}()
	<-c
	<-c
	return &su, &la
}

//IgnoreFileInFileList 用于排除不需要检测更新的文件或文件夹
func IgnoreFileInFileList(ignoreList *[]string, fileLists []*map[string]string) {
	var normIgnoreList []string
	for i := range *ignoreList {
		//(*ignoreList)[i] = filepath.FromSlash((*ignoreList)[i])
		normIgnoreList = append(normIgnoreList, filepath.FromSlash((*ignoreList)[i]))
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

//NormcaseFilelist 将目录名称中的slash转换为当前系统分隔符
func NormcaseFilelist(fileList *map[string]string) *map[string]string {
	normFileList := make(map[string]string)
	for k, v := range *fileList {
		normFileList[filepath.FromSlash(k)] = v
	}
	return &normFileList
}

//ToSlashFilelist 将目录名称中的分隔符转换为slash
func ToSlashFilelist(fileList *map[string]string) *map[string]string {
	slashFileList := make(map[string]string)
	for k, v := range *fileList {
		slashFileList[filepath.ToSlash(k)] = v
	}
	return &slashFileList
}

//AutoUpdate 全自动更新模式
func AutoUpdate(repair bool) {
	var newUpdateInfo updateInfo
	var localUpdateInfo updateInfo
	fmt.Println("正在获取版本信息")
	err := newUpdateInfo.LoadFromJSON(ReadStringFromURL(resourceURL + "update_info.json"))
	if err != nil {
		fmt.Println("网络异常或更新器版本过旧")
		os.Exit(1)
	}
	if IsFileOrDirectoryExists("update_info.json") {
		err = localUpdateInfo.LoadFromJSON(ReadStringFromFile("update_info.json"))
		if err != nil {
			fmt.Println("更新信息文件可能损坏\n请删除update_info.json")
			fmt.Println(err)
			if !repair {
				os.Exit(1)
			}
		}
	} else {
		localUpdateInfo = updateInfo{ResourceURL: resourceURL}
	}
	if localUpdateInfo.GameVersion < newUpdateInfo.GameVersion || repair { //版本旧了的话就更新
		if repair {
			fmt.Println("修复模式，忽略版本差异")
		}
		fmt.Printf("当前版本%d\n最新版本%d\n开始更新\n", localUpdateInfo.GameVersion, newUpdateInfo.GameVersion)
		fmt.Println("获取本地文件列表中")
		localFileList := GetFileList("game")
		fmt.Println("获取最新文件列表中")
		fileListJSON := ReadStringFromURL(resourceURL + "file_list.json")
		/*替换分隔符，现在用系统的自带方法，并强制规定json文件里都用slash(/)
		if runtime.GOOS == "windows" {
			fileListJSON = strings.Replace(fileListJSON, "/", `\\`, -1)
			for i := 0; i < len(newUpdateInfo.IgnoreList); i++ {
				newUpdateInfo.IgnoreList[i] = strings.Replace(newUpdateInfo.IgnoreList[i], "/", `\`, -1)
			}
		} else {
			fileListJSON = strings.Replace(fileListJSON, `\\`, "/", -1) //posix的符号替换
			for i := 0; i < len(newUpdateInfo.IgnoreList); i++ {
				newUpdateInfo.IgnoreList[i] = strings.Replace(newUpdateInfo.IgnoreList[i], `\`, "/", -1)
			}
		}*/
		newFileListUnformated := GetFileListFromJSON(fileListJSON)
		newFileList := NormcaseFilelist(newFileListUnformated) //规范化
		fmt.Println("对比文件差异中")
		surp, lack := CompareFileList(localFileList, newFileList)
		if !repair {
			IgnoreFileInFileList(&newUpdateInfo.IgnoreList, []*map[string]string{surp, lack})
		}
		for k := range *surp {
			fmt.Println("多余文件：" + k)
			fn := filepath.Join("rubbish", filepath.Base(k))
			if IsFileOrDirectoryExists(fn) {
				os.Remove(fn)
			}
			os.Rename(k, fn)
			fmt.Println("已移动至：" + fn)
		}
		if len(*lack) < 100 {
			for k := range *lack {
				fmt.Println("缺失文件：" + k)
			}
		} else {
			fmt.Printf("缺失文件数：%d\n", len(*lack))
		}
		//下载并更新文件
		nFile := len(*lack)          //需下载文件数
		limitor := make(chan int, 8) //限制了最多同时下载八个文件
		signal := make(chan int)
		for k, v := range *lack {
			go func(path, hash string) {
				limitor <- 0 //阻塞
				ok := false  //标记是否通过hash校验
				try := 0     //已尝试次数
				for !ok {
					if !IsFileOrDirectoryExists(filepath.Join("download", hash+".zip")) {
						DownLoadFile(resourceURL+"download/"+hash+".zip", filepath.Join("download", hash+".zip"))
					}
					Unzip(filepath.Join("download", hash+".zip"), filepath.Dir(path))
					ok = GetHash(path) == hash
					try++
					if !ok {
						os.Remove(path)
						os.Remove(filepath.Join("download", hash+".zip"))
						fmt.Println("下载或解压失败，重试中：" + hash)
					} else {
						fmt.Println("下载和解压成功" + hash)
						signal <- 0
					}
					if try > 5 {
						fmt.Println("下载或解压失败，超过最大重试次数：" + hash)
						signal <- 1
					}
				}
			}(k, v)
		}
		if nFile == 0 {
			close(signal)
		}
		succeed := 0 //统计结果
		failed := 0
		for sig := range signal { //等待下载完成
			nFile--
			<-limitor
			if sig == 0 {
				succeed++
			} else {
				failed++
			}
			fmt.Printf("已完成%d/%d\n", len(*lack)-nFile, len(*lack))
			if nFile == 0 {
				close(signal)
			}
		}
		close(limitor)
		fmt.Printf("更新完毕，%d成功，%d失败\n", succeed, failed)
		if failed == 0 {
			WriteStringToFile("update_info.json", newUpdateInfo.String())
		}
	} else {
		fmt.Println("已是最新版，如需修复游戏文件请附加参数--repair")
	}
}

//Pack 制作更新包
func Pack() {
	fmt.Println("开始制作更新包")
	os.MkdirAll(filepath.Join("package", "download"), os.ModePerm)
	fileList := GetFileList("game")
	fmt.Println("已获得文件列表")
	fileListJSON, _ := json.Marshal(ToSlashFilelist(fileList))
	WriteStringToFile(filepath.Join("package", "file_list.json"), string(fileListJSON))
	fmt.Println("已写入文件列表")
	nFiles := len(*fileList)
	c := make(chan int)
	for k, v := range *fileList {
		go func(path, hash string) {
			os.Remove(filepath.Join("package", "download", hash+".zip"))
			Zip(path, filepath.Join("package", "download", hash+".zip"))
			fmt.Println("已压缩：" + path)
			c <- 0
		}(k, v)
	}
	for range c {
		nFiles--
		if nFiles == 0 {
			close(c)
		}
	}
	fmt.Println("制包完毕，记得更新update_info.json中的版本号")
}

//Repair 游戏文件修复
func Repair() {
	AutoUpdate(true)
}

func main() {
	fmt.Println("Minecraft自动更新器启动")
	requiredDir := []string{"game", "rubbish", "download"}
	for _, rDirName := range requiredDir {
		if !IsFileOrDirectoryExists(rDirName) {
			os.Mkdir(rDirName, os.ModePerm)
		}
	}
	if len(os.Args) == 1 {
		fmt.Println("自动模式\n如需知晓更多功能请附加参数--help")
		AutoUpdate(false)
		os.Exit(0)
	}
	if len(os.Args) > 1 {
		if os.Args[1] == "--help" {
			fmt.Println("--help    获取帮助\n--pack    制作更新包\n--repair  游戏文件修复模式\n不附加参数即为自动模式")
		}
		if os.Args[1] == "--pack" {
			Pack()
		}
		if os.Args[1] == "--repair" {
			Repair()
		}
		os.Exit(0)
	}
	fmt.Println("未知参数:", os.Args[1])
	fmt.Println("附加--help参数以获取帮助")
	os.Exit(1)
}