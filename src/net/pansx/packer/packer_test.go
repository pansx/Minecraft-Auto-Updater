package packer

import "testing"

func TestPacker(t *testing.T) {
	Pack(false)
}
func TestUploadAll(t *testing.T) {
	UploadAll()
}

func TestFtpInfo_RenameToDownload(t *testing.T) {
	ftpInfo := New("ftp.json")
	ftpInfo.RenameToDownload()
}
