package libs

import (
	"mime"
	"os"
)

// 获取文件的基本信息
func GetFilestat(file string) (filestat os.FileInfo, err error) {
	filestat,err = os.Stat(file)
	if err != nil {
		return
	}

	return filestat,err
}

// 根据文件后缀（扩展名）获取文件的mimeType
func GetFileMimeType(typename string) (mediatype string, err error) {
	if len(typename) < 1 {
		mediatype = ""
		err = nil
		return;
	}
	mediatype, _, err = mime.ParseMediaType(typename)
	if err!= nil {
		return;
	}
	return mediatype,nil
}
