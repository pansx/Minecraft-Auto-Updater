package main

import (
	"../utils"
	"fmt"
	"github.com/shettyh/threadpool"
	"io"
	"net/http"
	"os"
	"strings"
)

type Downloader struct {
	workerNum int
	destDir   string
	results   []*threadpool.Future
	pool      *threadpool.ThreadPool
}

type DownloadCallable struct {
	url     string
	method  string
	hash    string
	destDir string
}

func New(destDir string) Downloader {
	d := Downloader{}
	d.workerNum = 8
	d.destDir = destDir
	d.pool = threadpool.NewThreadPool(d.workerNum, 9999999)
	return d
}

func (d *DownloadCallable) Call() interface{} {
	//Do task
	result := 1
	if d.method == "downloadAndCheck" {
		err := d.DownLoadFileAndCheck(d.url, d.hash)
		if err != nil {
			result = 0
		}
	} else if d.method == "downlod" {
		err := d.DownLoadFile(d.url)
		if err != nil {
			result = 0
		}
	}
	return result
}

func main() {

	// 为了使用 downloadWorker 线程池并且收集他们的结果，我们需要 2 个通道。
	downloader := New("R:\\pansx\\OneDrive\\project\\Java\\Minecraft-Auto-Updater\\dist")
	// 这里我们发送 9 个 `jobs`，然后 `close` 这些通道
	// 来表示这些就是所有的任务了。
	urls := []string{"123"}
	for i := 0; i < 100; i++ {
		urls = append(urls, utils.GenUUID())
	}
	urls = append(urls, "321")
	fmt.Println(urls, downloader)
	downloader.setDownloadQueue(urls)
	result := downloader.startDownloadUntilGetResult()

	fmt.Println(result)

}

func (d *Downloader) setDownloadQueue(urls []string) {
	for _, url := range urls {
		callable := &DownloadCallable{url: url, method: "download"}
		future, err := d.pool.ExecuteFuture(callable)
		if err != nil {
			fmt.Println(err)
		}
		d.results = append(d.results, future)
	}
}
func (d *Downloader) startDownloadUntilGetResult() []int {
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
	if utils.IsFileOrDirectoryExists(d.destDir) {
		return nil
	}
	destFile, err := os.Create(d.destDir)
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
func (d *DownloadCallable) DownLoadFileAndCheck(url, hash string) error {
	//log.Println("下载文件并检查 url:" + url + " dest:" + destDir)
	hash = strings.ToLower(hash)
	if utils.IsFileOrDirectoryExists(d.destDir) {
		if hash == utils.GetHash(d.destDir) {
			//log.Println("文件校验通过 url:" + url + " dest:" + destDir)
			return nil
		}
		os.Remove(d.destDir)
		//log.Println("文件校验不通过，重新下载 url:" + url + " dest:" + destDir)
	}
	for i := 0; ; i++ {
		err := d.DownLoadFile(url)
		if err == nil {
			if hash == utils.GetHash(d.destDir) {
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
