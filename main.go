package main

import (
	"fmt"
	"html/template"
	"image-server/libs"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	UPLOAD_DIR = "uploads/v1"
	ListDir    = 0x0001
)

func main() {
	const PORT = 8005
	http.HandleFunc("/favicon.ico", safeHandler(showFavicon))
	http.HandleFunc("/", safeHandler(listHandler))
	http.HandleFunc("/upload", safeHandler(uploadHandler))
	http.HandleFunc("/view", safeHandler(viewHandler))
	fmt.Println("http server in " + strconv.Itoa(PORT) + ".")
	err := http.ListenAndServe(":"+strconv.Itoa(PORT), nil)
	if err != nil {
		fmt.Println(err)
	}
}

//静态资源服务器
func staticDirHandler(mux *http.ServeMux, prefix, staticDir string, flags int) {
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		file := staticDir + r.URL.Path[len(prefix)-1:]
		if (flags & ListDir) == 0 {
			http.NotFound(w, r)
		}
		http.ServeFile(w, r, file)
	})
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		renderHtml(w, "views/upload.html", nil)
		return
	} else if r.Method == "POST" {
		f, h, err := r.FormFile("image")
		checkError(err)
		filename := h.Filename
		defer f.Close()
		t, err := os.Create(UPLOAD_DIR + "/" + filename)
		defer t.Close()
		_, err = io.Copy(t, f)
		checkError(err)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	fileInfoArr, err := ioutil.ReadDir(UPLOAD_DIR)
	checkError(err)
	var locals = make(map[string]interface{})
	images := []string{}
	for _, fileInfo := range fileInfoArr {
		imageId := fileInfo.Name()
		images = append(images, imageId)
	}
	locals["images"] = images
	err = renderHtml(w, "views/list.html", locals)
	checkError(err)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	imageId := r.FormValue("id")
	imagePath := UPLOAD_DIR + "/" + imageId
	if exists := fileExists(imagePath); !exists {
		http.NotFound(w, r)
	}
	// 判断文件的后缀是否是图片文件，如果不是图片则直接下载
	_,fileName := filepath.Split(imagePath)
	fileExt := path.Ext(imagePath)
	mediatype, err := libs.GetFileMimeType(fileExt)
	checkError(err)
	if strings.Contains(imagePath, ".gitignore") {
		w.Header().Set("Content-type", "text/html")
	} else if strings.Contains(mediatype,"image") {
		w.Header().Set("Content-type", mediatype)
	} else {
		// 不是图片文件，则进行下载操作
		w.Header().Set("Content-type", "text/plain")
		w.Header().Set("Content-Disposition", "attachment;fileName="+fileName)
	}

	http.ServeFile(w, r, imagePath)
}

func renderHtml(w http.ResponseWriter, path string, locals map[string]interface{}) error {
	t, err := template.ParseFiles(path)
	if err != nil {
		return err
	}
	return t.Execute(w, locals)
}

func safeHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err!= nil {
				if errStr, ok := err.(string); !ok {
					http.Error(w, errStr, http.StatusInternalServerError)
					log.Printf("WARN:panic in %#v-%#v \n", fn, errStr)// %#v ---> a Go-syntax representation of the value
				}
			}
		}()
		fn(w, r)
	}
}

func showFavicon(w http.ResponseWriter, r *http.Request)  {
	imagePath := "favicon.ico"
	if exists := fileExists(imagePath); !exists {
		http.NotFound(w, r)
	}
	w.Header().Set("Content-type", "image/png")
	http.ServeFile(w, r, imagePath)
}

func checkError(err error) {
	if err != nil {
		panic(err)
		return
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	return os.IsExist(err)
}
