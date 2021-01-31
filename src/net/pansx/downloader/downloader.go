package downloader

import (
	"errors"
	"fmt"
	"github.com/shettyh/threadpool"
	"io"
	"net/http"
	"net/pansx/fileInfo"
	"net/pansx/utils"
	"os"
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

func New(destDir string, host string) Downloader {
	d := Downloader{}
	d.workerNum = 32
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
	} else if d.method == "download" {
		err := d.DownLoadFile(d.url)
		if err != nil {
			fmt.Println(err)
			result = 0
		} else {
			fmt.Println("下载完毕:", d.downloadFile)
		}
	}
	return result
}

func (d *Downloader) SetDownloadQueue(fiList []*fileInfo.FileInfo) {
	for _, fi := range fiList {
		callable := &DownloadCallable{
			url:          d.host + "download/" + fi.Name,
			method:       "download",
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
		ints = append(
			ints,
			get.(int),
		)
	}
	return ints
}

//DownLoadFile 下载文件
func (d *DownloadCallable) DownLoadFile(url string) error {
	if utils.IsFileOrDirectoryExists(d.downloadFile) {
		return nil
	}

	res, err := http.Get(url)

	if err != nil || res.StatusCode != 200 {
		return errors.New("下载失败!" + url)
	}
	destFile, err := os.Create(d.downloadFile)
	defer destFile.Close()
	_, err = io.Copy(destFile, res.Body)
	_ = destFile.Close()
	err = utils.Unzip(d.downloadFile, d.destFile)
	return err
}

//DownLoadFileAndCheck 下载文件并校验hash是否相符
func (d *DownloadCallable) DownLoadFileAndCheck(url, hash string) error {
	//log.Println("下载文件并检查 url:" + url + " dest:" + downloadFile)
	hash = strings.ToLower(hash)
	if utils.IsFileOrDirectoryExists(d.destFile) {
		if hash == utils.GetHash(d.destFile) {
			//log.Println("文件校验通过 url:" + url + " dest:" + downloadFile)
			return nil
		}
		os.Remove(d.destFile)
		//log.Println("文件校验不通过，重新下载 url:" + url + " dest:" + downloadFile)
	}
	for i := 0; ; i++ {
		err := d.DownLoadFile(url)
		if err == nil {
			if hash == utils.GetHash(d.downloadFile) {
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
