package updateInfo

import (
	"encoding/json"
	"net/pansx/utils"
)

type UpdateInfo struct {
	GameVersion       int         `json:"version"`
	Mirror            string      `json:"mirror"`
	IgnoreList        []string    `json:"ignore_list"`
	PackageIgnoreList []string    `json:"package_ignore_list"`
	FileInfoList      []*FileInfo `json:"file_Info_list"`
	//CommandToRun      string   `json:"command_to_run"`
}
type FileInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Hash string `json:"hash"`
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
	ui.FileInfoList = []*FileInfo{}
	for s, v := range m {
		ui.FileInfoList = append(ui.FileInfoList, &FileInfo{
			Name: utils.GenUUID() + ".zip",
			Path: s,
			Hash: v,
		})
	}
}
