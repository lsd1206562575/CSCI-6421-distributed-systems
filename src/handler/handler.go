package handler

import (
	"dfs/meta"
	utils "dfs/util"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

// 文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data, err := ioutil.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "Interal server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Print("Failed to get data, err:%s", err.Error())
			return
		}
		defer file.Close()

		fileMeta := meta.FileMeta{
			FileName:  head.Filename,
			Location:  "/tmp/" + head.Filename,
			UpdatedAt: time.Now().Format("2006-01-02 15:04"),
		}

		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Print("Failed to create file, err:%s", err.Error())
			return
		}
		defer newFile.Close()

		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Print("Failed to save data into file, err:%s", err.Error())
			return
		}

		newFile.Seek(0, 0)
		fileMeta.FileSha1 = utils.FileSha1(newFile)
		meta.UpdateFileMeta(fileMeta)

		InitialMultipartUploadHandler(fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)

		http.Redirect(w, r, "upload/suc", http.StatusFound)
	}
}

// 上传完成
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Upload finished!")
}

// 获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	fMeta := meta.GetFileMeta(filehash)
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)
	f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Description", "attachment; filename=\""+fm.FileName+"\"")
	w.Write(data)

}
