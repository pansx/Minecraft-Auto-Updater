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
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

const resourceURL = "https://minecraft-updater.oss-cn-shanghai.aliyuncs.com/"

//Zip 用于压缩文件srcFile可以是单文件也可以是目录
func Zip(srcFile string, destZip string) error {
	archive, err := createZip(destZip)
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
		/*
			替换分隔符，现在用自带方法
			sig := "/"
			if runtime.GOOS == "windows" {
				sig = `\`
			}
			header.Name = strings.TrimPrefix(path, filepath.Dir(srcFile)+sig)
		*/
		//header.Name = strings.TrimPrefix(path, filepath.FromSlash(filepath.ToSlash(filepath.Dir(srcFile)+"/")))
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

func createZip(destZip string) (*zip.Writer, error) {
	zipfile, err := os.Create(destZip)
	if err != nil {
		return nil, err
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	return archive, err
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

type updateInfo struct {
	GameVersion       int      `json:"version"`
	IgnoreList        []string `json:"ignore_list"`
	PackageIgnoreList []string `json:"package_ignore_list"`
	//CommandToRun      string   `json:"command_to_run"`
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

//LaunchGameLauncher 不翻译了
func LaunchGameLauncher(output io.Writer) {
	fmt.Fprintln(output, "正在启动游戏启动器，请不要关闭更新器窗口")
	cmd := exec.Command("java", "-jar", "Launcher.jar")
	cmd.Dir = "./game"
	//cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		fmt.Fprintln(output, "启动失败，你可能没安装java")
		time.Sleep(60 * time.Second)
	}
}

//AutoUpdate 全自动更新模式
func AutoUpdate(repair bool, output io.Writer) {
	var newUpdateInfo updateInfo
	var localUpdateInfo updateInfo
	fmt.Fprintln(output, "正在获取版本信息")
	err := newUpdateInfo.LoadFromJSON(ReadStringFromURL(resourceURL + "update_info.json"))
	if err != nil {
		fmt.Fprintln(output, "网络异常或更新器版本过旧")
		os.Exit(1)
	}
	if IsFileOrDirectoryExists("update_info.json") {
		err = localUpdateInfo.LoadFromJSON(ReadStringFromFile("update_info.json"))
		if err != nil {
			fmt.Fprintln(output, "更新信息文件可能损坏\n请删除update_info.json")
			fmt.Fprintln(output, err)
			if !repair {
				os.Exit(1)
			}
		}
	} else {
		localUpdateInfo = updateInfo{}
	}
	if localUpdateInfo.GameVersion < newUpdateInfo.GameVersion || repair { //版本旧了的话就更新
		if repair {
			fmt.Fprintln(output, "修复模式，忽略版本差异")
		}
		fmt.Fprintf(output, "当前版本:%d 最新版本:%d\n", localUpdateInfo.GameVersion, newUpdateInfo.GameVersion)
		//fmt.Println("获取本地文件列表中")
		localFileList := GetFileList("game")
		//fmt.Println("获取最新文件列表中")
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
		//fmt.Println("对比文件差异中")
		surp, lack := CompareFileList(localFileList, newFileList)
		if !repair {
			IgnoreFileInFileList(&newUpdateInfo.IgnoreList, []*map[string]string{surp, lack}, true)
		}
		for k := range *surp {
			//fmt.Println("多余文件：" + k)
			fn := filepath.Join("rubbish", filepath.Base(k))
			if IsFileOrDirectoryExists(fn) {
				os.Remove(fn)
			}
			os.Rename(k, fn)
			//fmt.Println("已移动至：" + fn)
		}
		fmt.Fprintf(output, "多余文件数:%d 缺失文件数:%d\n", len(*surp), len(*lack))
		//下载并更新文件
		nFile := len(*lack)           //需下载文件数
		limitor := make(chan int, 64) //限制了最多同时下载64个文件
		signal := make(chan int)
		hashToShow := make(chan string)
		for k, v := range *lack {
			go func(path, hash string) {
				limitor <- 0 //阻塞
				ok := false  //标记是否通过hash校验
				try := 0     //已尝试次数
				conflictSuffix := ""
				for !ok {
					if !IsFileOrDirectoryExists(filepath.Join("download", hash+".zip"+conflictSuffix)) {
						DownLoadFile(resourceURL+"download/"+hash+".zip"+conflictSuffix, filepath.Join("download", hash+".zip"+conflictSuffix))
					}
					Unzip(filepath.Join("download", hash+".zip"+conflictSuffix), filepath.Dir(path))
					ok = GetHash(path) == hash
					try++
					if !ok {
						conflictSuffix += ".conflict"
						/*
							if !IsFileOrDirectoryExists(path) {
								conflictSuffix += ".conflict"
							} else {
								os.Remove(path)
							}
							os.Remove(filepath.Join("download", hash+".zip"+conflictSuffix))
						*/
						//fmt.Println("下载或解压失败，重试中：" + hash)
					} else {
						//fmt.Println("下载和解压成功" + hash)
						signal <- 0
						hashToShow <- hash
						return
					}
					if try > 15 {
						//fmt.Println("下载或解压失败，超过最大重试次数：" + hash)
						fmt.Println("文件更新超过最大重试次数：" + path)
						signal <- 1
						hashToShow <- hash
						return
					}
				}
			}(k, v)
		}
		if nFile == 0 {
			close(signal)
		}
		succeed := 0 //统计结果
		failed := 0
		_, isView := output.(*gocui.View)
		for sig := range signal { //等待下载完成
			nFile--
			<-limitor
			if sig == 0 {
				succeed++
			} else {
				failed++
			}
			h := <-hashToShow
			if isView && ((len(*lack)-nFile)%int(float32(len(*lack))*0.1) == 0) {
				fmt.Fprintf(output, "已完成[%v/%v]:%v\n", len(*lack)-nFile, len(*lack), h)
			} else {
				fmt.Fprintf(output, "已完成[%v/%v]:%v\r", len(*lack)-nFile, len(*lack), h)
			}
			if nFile == 0 {
				close(signal)
				if !isView {
					fmt.Fprintln(output, "")
				}
			}
		}
		close(limitor)
		fmt.Fprintf(output, "成功:%d 失败:%d\n", succeed, failed)
		if failed == 0 {
			WriteStringToFile("update_info.json", newUpdateInfo.String())
			fmt.Fprintf(output, "更新成功\n")
			LaunchGameLauncher(output)
			return
		}
		fmt.Fprintln(output, "更新失败，请重新运行更新器")
		time.Sleep(60 * time.Second)
	} else {
		fmt.Fprintln(output, "已是最新版")
		LaunchGameLauncher(output)
	}
}

//Pack 制作更新包如果 init用于表明是否是第一次打包，如果不是，更新器就会从服务端获取旧文件列表
func Pack(init bool) {
	fmt.Println("开始制作更新包")
	os.RemoveAll(filepath.Join("package", "download"))
	os.MkdirAll(filepath.Join("package", "download"), os.ModePerm)
	fileList := GetFileList("game")
	fileList = ToSlashFilelist(fileList)
	fmt.Println("已获得文件列表")
	if init {
		var initUpdateInfo updateInfo
		initUpdateInfo.GameVersion = 0
		initUpdateInfo.IgnoreList = []string{}
		initUpdateInfo.PackageIgnoreList = []string{}
		initUpdInfoJSON, _ := json.Marshal(initUpdateInfo)
		WriteStringToFile(filepath.Join("package", "update_info.json"), string(initUpdInfoJSON))
		fileListJSON, _ := json.Marshal(ToSlashFilelist(fileList))
		WriteStringToFile(filepath.Join("package", "file_list.json"), string(fileListJSON))
		println("初始化已完成，请将完善后的update_info.json和file_list.json上传到文件服务器，然后再进行打包")
		return
	}
	var newUpdateInfo updateInfo
	err := newUpdateInfo.LoadFromJSON(ReadStringFromURL(resourceURL + "update_info.json"))
	if err != nil {
		fmt.Println("读取更新信息时出错，是不是忘了初始化文件服务器？输入--init以获取初始化所需的文件，输入--help获取更多帮助")
		return
	}
	newUpdateInfo.GameVersion++ //自动修改游戏版本
	_ = os.Remove(filepath.Join("package", "update_info.json"))
	newUpdateInfoJSON, _ := json.Marshal(newUpdateInfo)
	WriteStringToFile(filepath.Join("package", "update_info.json"), string(newUpdateInfoJSON)) //写入修改了游戏版本的更新信息文件
	oldFileList := GetFileListFromJSON(ReadStringFromURL(resourceURL + "file_list.json"))      //服务器上的文件列表
	oldFileList = ToSlashFilelist(oldFileList)
	IgnoreFileInFileList(&newUpdateInfo.PackageIgnoreList, []*map[string]string{fileList, oldFileList}, false) //有一些文件不打包
	/*由于文件重名问题，暂时停用增量压缩功能
	surp, _ := CompareFileList(fileList, oldFileList)                                                          //多出来的文件要另外打包
	os.RemoveAll(filepath.Join("package", "download_surp"))
	os.MkdirAll(filepath.Join("package", "download_surp"), os.ModePerm)
	nFiles := len(*surp)
	nFilesTotal := nFiles
	if nFiles != 0 {
		c := make(chan string)
		for k, v := range *surp {
			go func(path, hash string) {
				Zip(path, filepath.Join("package", "download_surp", hash+".zip"))
				c <- path
			}(k, v)
		}
		for p := range c {
			nFiles--
			fmt.Fprintf(os.Stdout, "增量压缩中[%v/%v]:%v\n", nFilesTotal-nFiles, nFilesTotal, p)
			if nFiles == 0 {
				close(c)
				fmt.Println("")
			}
		}
	}
	*/
	os.Remove(filepath.Join("package", "file_list.json"))
	fileListJSON, _ := json.Marshal(ToSlashFilelist(fileList))
	WriteStringToFile(filepath.Join("package", "file_list.json"), string(fileListJSON))
	fmt.Println("已写入文件列表")
	nFiles := len(*fileList)
	nFilesTotal := nFiles
	c := make(chan string)
	limitor := make(chan int)
	for k, v := range *fileList {
		go func(path, hash string, limitor chan int) {
			<-limitor
			destZip := filepath.Join("package", "download", hash+".zip")
			for IsFileOrDirectoryExists(destZip) { //解决同名文件冲突
				fmt.Printf("遇到冲突，尝试解决中...%s\n", destZip)
				destZip += ".conflict"
			}
			Zip(path, destZip)
			c <- path
		}(k, v, limitor)
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

//Repair 游戏文件修复
func Repair() {
	AutoUpdate(true, os.Stdout)
}

func main() {
	fmt.Println("======Minecraft自动更新器======")
	requiredDir := []string{"game", "rubbish", "download"}
	for _, rDirName := range requiredDir {
		if !IsFileOrDirectoryExists(rDirName) {
			os.Mkdir(rDirName, os.ModePerm)
		}
	}
	if len(os.Args) == 1 {
		//fmt.Println("自动模式\n如需知晓更多功能请附加参数--help")
		AutoUpdate(false, os.Stdout)
		//cui()
		//os.Exit(0)
	}
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--help":
			fmt.Println("--help    获取帮助\n--init    初次制作更新包\n--pack    制作更新包\n--repair  游戏文件修复模式\n不附加参数即为自动模式")
		case "--pack":
			Pack(false)
		case "--init":
			Pack(true)
		case "--repair":
			Repair()
		case "--debug":
			//cui()
		default:
			fmt.Println("未知参数:", os.Args[1])
			fmt.Println("附加--help参数以获取帮助")
			os.Exit(1)
		}
		os.Exit(0)
	}
}
