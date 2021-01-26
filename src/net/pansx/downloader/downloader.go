package main

import (
	"../utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)
import "time"

type Downloader struct {
	destDir     string
	jobs        chan string
	results     chan int
	workerNum   int
	queueLength int
}

func New(destDir string) Downloader {
	d := Downloader{}
	d.destDir = destDir
	d.jobs = make(chan string, 100)
	d.results = make(chan int, 100)
	d.workerNum = 8
	d.queueLength = 0
	d.initWorker()
	return d
}

func (d *Downloader) initWorker() {
	for w := 1; w <= d.workerNum; w++ {
		go d.downloadWorker(w)
	}
}

func (d *Downloader) downloadWorker(id int) {
	for j := range d.jobs {
		fmt.Println("downloadWorker", id, "processing job", j)
		time.Sleep(time.Second)
		d.results <- 1
	}
}

func main() {

	// 为了使用 downloadWorker 线程池并且收集他们的结果，我们需要 2 个通道。
	downloader := New("123")
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
	d.results = make(chan int, len(urls))
	for _, url := range urls {
		d.jobs <- url
		d.queueLength += 1
	}
	close(d.jobs)
}
func (d *Downloader) startDownloadUntilGetResult() []int {
	var ints []int
	for i := 0; i < d.queueLength; i++ {
		res := <-d.results
		ints = append(ints, res)
	}
	return ints
}

//DownLoadFile 下载文件
func (d *Downloader) DownLoadFile(url string) error {
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
func (d *Downloader) DownLoadFileAndCheck(url, hash string) error {
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
