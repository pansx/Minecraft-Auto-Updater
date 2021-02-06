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
	var fileInfoList []*fileInfo.FileInfo
	for s, v := range m {
		fileInfoList = append(fileInfoList, &fileInfo.FileInfo{
			Name: ui.getName(s, v),
			Path: s,
			Hash: v,
		})
	}
	ui.FileInfoList = fileInfoList
}

func (ui *UpdateInfo) getName(path, hash string) string {
	for _, info := range ui.FileInfoList {
		if info.Path == path && hash == hash {
			return info.Name
		}
	}
	newName := utils.GenUUID() + ".zip"
	return newName
}
func (ui *UpdateInfo) CheckAndDownloadAll(destPath string) []int {
	d := downloader.New(destPath, ui.Mirror)
	infos := []*fileInfo.FileInfo{{
		Name: "ui.zip",
		Path: "upd/ui/ui.exe",
		Hash: "ca025c954216ff02eb4bf37fc0df5159753e40ee",
	}}
	d.SetDownloadQueue(append(infos, ui.FileInfoList...))
	result := d.StartDownloadUntilGetResult()
	return result
}
func (ui *UpdateInfo) MakeNewFileInfo() {
	fmt.Println("开始获得文件列表,这可能需要几分钟...")
	fileList := utils.GetFileHashList("game")
	ui.LoadFileInfoByMap(fileList)
	fmt.Println("已获得文件列表")
}
