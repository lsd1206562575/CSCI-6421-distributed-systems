package handler

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"io"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	rPool "dfs/cache/redis"
)

type MultipartUploadInfo struct {
	FileHash   string
	FileSize   int
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

func InitialMultipartUploadHandler(w http.ResponseWriter, r *http.Request) {
	//解析用户请求参数
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		return
	}

	//获得redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//生成分块上传的初始化信息
	upinfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   fmt.Sprint("%x", time.Now().UnixNano()),
		ChunkSize:  64 * 1024 * 1024,
		ChunkCount: int(math.Ceil(float64(filesize) / (64 * 1024 * 1024))),
	}

	//将初始信息写入redis缓存
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "chunkcount", upinfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filehash", upinfo.FileHash)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filesize", upinfo.FileSize)

	//响应初始化数据返回客户端
	io.WriteString(w, "Success")

}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	//解析用户请求参数
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")

	//获得redis连接池中的连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//获得文件句柄， 用于存储分块内容
	fpath := "/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0777)
	fd, err := os.Create(fpath)
	if err != nil {
		w.WriteHeader(401)
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	//更新redis缓存状态
	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
	//返回处理结果
	io.WriteString(w, "Success")
}

// 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	//解析请求参数
	r.ParseForm()
	upid := r.Form.Get("uploadid")
	//filehash := r.Form.Get("filehash")
	//filesize := r.Form.Get("filesize")
	//filename := r.Form.Get("filename")

	//获得redis连接池中的连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//通过uploadid查询redis并判断所有分块是否上传完成
	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upid))
	if err != nil {
		return
	}
	totalCount := 0
	chunkCount := 0

	for i := 0; i < len(data); i += 2 {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount += 1
		}
	}

	if totalCount != chunkCount {
		return
	}

	//合并分块

	//更新唯一文件表

	//响应处理结果

}
