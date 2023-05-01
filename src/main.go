package main

import (
	"dfs/handler"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/file/upload", handler.UploadHandler)
	http.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
	http.HandleFunc("/file/meta", handler.GetFileMetaHandler)
	http.HandleFunc("/file/download", handler.DownloadHandler)

	//分块上传接口
	http.HandleFunc("/file/mpupload/init", handler.InitialMultipartUploadHandler)
	http.HandleFunc("/file/mpupload/uppart", handler.InitialMultipartUploadHandler)
	http.HandleFunc("/file/mpupload/complete", handler.InitialMultipartUploadHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Print("Failed to start server, err:%s", err.Error())
	}
}
