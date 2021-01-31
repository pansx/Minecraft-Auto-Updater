package updateInfo

import (
	"encoding/json"
	"fmt"
	"net/pansx/downloader"
	"net/pansx/fileInfo"
	"net/pansx/utils"
	"net/url"
)

type UpdateInfo struct {
	GameVersion       int                  `json:"version"`
	Mirror            string               `json:"mirror"`
	IgnoreList        []string             `json:"ignore_list"`
	PackageIgnoreList []string             `json:"package_ignore_list"`
	FileInfoList      []*fileInfo.FileInfo `json:"file_Info_list"`
	//CommandToRun      string   `json:"command_to_run"`
}

func New(path string) *UpdateInfo {
	u := &UpdateInfo{}
	url, err := url.Parse(path)
	if err != nil || url.Host == "" {
		_ = u.LoadFromJSON(utils.ReadStringFromFile(path))
	} else {
		urlString := url.String()
		_ = u.LoadFromJSON(utils.ReadStringFromURL(urlString))
	}
	return u
}

func getFromUrl(url string) *UpdateInfo {
	u := &UpdateInfo{}
	u.LoadFromJSON(utils.ReadStringFromURL(url))
	return u
}

//返回的是json形式的表达
func (ui *UpdateInfo) String() string {

	d, _ := json.Marshal(ui)
	return string(d)
}

//LoadFromJSON 从json中加载updateInfo
func (ui *UpdateInfo) LoadFromJSON(s string) error {
	err := json.Unmarshal([]byte(s), ui)
	return err
}

func (ui *UpdateInfo) LoadFileInfoByMap(m map[string]string) {
	ui.FileInfoList = []*fileInfo.FileInfo{}
	for s, v := range m {
		ui.FileInfoList = append(ui.FileInfoList, &fileInfo.FileInfo{
			Name: utils.GenUUID() + ".zip",
			Path: s,
			Hash: v,
		})
	}
}
func (ui *UpdateInfo) CheckAndDownloadAll(destPath string) error {
	d := downloader.New(destPath, ui.Mirror)
	d.SetDownloadQueue(ui.FileInfoList)
	result := d.StartDownloadUntilGetResult()
	fmt.Println(result)
	return nil
}
