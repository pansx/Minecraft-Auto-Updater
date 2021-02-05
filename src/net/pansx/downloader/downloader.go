package downloader

import (
	"errors"
	"fmt"
	"github.com/shettyh/threadpool"
	"io"
	"log"
	"net/http"
	"net/pansx/fileInfo"
	"net/pansx/utils"
	"os"
	"os/exec"
	"path"
	"strings"
)

type Downloader struct {
	workerNum int
	destDir   string
	host      string
	results   []*threadpool.Future
	pool      *threadpool.ThreadPool
}

type DownloadCallable struct {
	url          string
	method       string
	hash         string
	downloadFile string
	destFile     string
}

var progress = 0

func New(destDir string, host string) Downloader {
	_ = exec.Command("cmd", "/c", "title", fmt.Sprintf("mcupd-%p", &progress)).Run()
	d := Downloader{}
	d.workerNum = 8
	d.destDir = destDir
	d.host = host
	d.pool = threadpool.NewThreadPool(d.workerNum, 9999999)
	return d
}

func (d *DownloadCallable) Call() interface{} {
	//Do task
	result := 1
	fmt.Println("下载开始:", d.downloadFile)
	if d.method == "downloadAndCheck" {
		err := d.DownLoadFileAndCheck(d.url, d.hash)
		if err != nil {
			result = 0
		}
		d.execUi()
	} else if d.method == "download" {
		err := d.DownLoadFile(d.url)
		if err != nil {
			fmt.Println(err)
			result = 0
		} else {
			fmt.Println("下载完毕:", d.destFile)
		}
	}
	return result
}

func (d *Downloader) SetDownloadQueue(fiList []*fileInfo.FileInfo) {
	for _, fi := range fiList {
		callable := &DownloadCallable{
			url:          d.host + "download/" + fi.Name,
			method:       "downloadAndCheck",
			downloadFile: path.Join(d.destDir, fi.Name),
			destFile:     fi.GetDeployPath(),
			hash:         fi.Hash,
		}
		future, err := d.pool.ExecuteFuture(callable)
		if err != nil {
			fmt.Println(err)
		}
		d.results = append(d.results, future)
	}
}
func (d *Downloader) StartDownloadUntilGetResult() []int {
	var ints []int

	for _, result := range d.results {
		get := result.Get()
		progress++
		ints = append(
			ints,
			get.(int),
		)
		fmt.Println("多线程下载进度:", progress, "/", len(d.results))
	}
	return ints
}

//DownLoadFile 下载文件
func (d *DownloadCallable) DownLoadFile(url string) error {
	remoteLength, err := d.getRemoteLength(url)
	if !d.isFileExistAndValid(url, remoteLength) {
		res, err := http.Get(url)
		if err != nil || res.StatusCode != http.StatusOK {
			return errors.New("下载失败!" + url)
		}
		destFile, err := os.Create(d.downloadFile)
		defer destFile.Close()
		_, err = io.Copy(destFile, res.Body)
		stat, _ := destFile.Stat()
		if stat.Size() != remoteLength {
			fmt.Println("下载的文件与远程声称的文件大小不一致!重试", stat.Size(), remoteLength, url)
			_ = destFile.Close()
			_ = os.Remove(d.downloadFile)
			return d.DownLoadFile(url)
		}
		_ = destFile.Close()
	}
	err = utils.Unzip(d.downloadFile, d.destFile)
	return err
}

func (d *DownloadCallable) execUi() {
	uiPath := "upd/ui/ui.exe"
	if d.destFile == uiPath {
		_ = exec.Command(uiPath).Start()
	}
}

func (d *DownloadCallable) getRemoteLength(url string) (int64, error) {
	head, err := http.Head(url)
	if head == nil || head.ContentLength == 0 || head.StatusCode != http.StatusOK {
		err = errors.New("无法获得文件信息,下载失败!" + url)
	}
	remoteLength := head.ContentLength
	return remoteLength, err
}

func (d *DownloadCallable) isFileExistAndValid(url string, remoteLength int64) bool {
	if utils.IsFileOrDirectoryExists(d.downloadFile) {
		f, err := os.Open(d.downloadFile)
		stat, err := f.Stat()
		if stat.Size() == remoteLength {
			return true
		}
		if err != nil {
			fmt.Println("检测到服务端文件和本地已存在的文件大小不一致,重新获取...", url)
			_ = os.Remove(d.downloadFile)
		}
	}
	return false
}

//DownLoadFileAndCheck 下载文件并校验hash是否相符
func (d *DownloadCallable) DownLoadFileAndCheck(url, hash string) error {
	//log.Println("下载文件并检查 url:" + url + " dest:" + downloadFile)
	hash = strings.ToLower(hash)
	if utils.IsFileOrDirectoryExists(d.destFile) {
		getHash := utils.GetHash(d.destFile)
		if hash == getHash {
			//log.Println("文件校验通过 url:" + url + " dest:" + downloadFile)
			return nil
		}
		os.Remove(d.destFile)
		log.Println("文件校验不通过，重新下载 url:" + url + " dest:" + url)
	}
	for i := 0; ; i++ {
		err := d.DownLoadFile(url)
		if err == nil {
			if hash == utils.GetHash(d.destFile) {
				//log.Println("文件校验通过 url:" + url + " dest:" + downloadFile)
				return nil
			}
		} else {
			if i > 3 { //最多尝试十次
				//log.Println("超过最大重试次数 url:" + url + " dest:" + downloadFile)
				return err
			}
		}
	}
}
