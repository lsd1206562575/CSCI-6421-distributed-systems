package handler

import (
	utils "dfs/util"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"path"
	"strconv"
	"time"

	rPool "dfs/cache/redis"
)

type MultipartUploadInfo struct {
	FileHash   string
	FileName   string
	FileSize   int64
	UploadID   string
	ChunkSize  int
	ChunkCount int
}

func InitialMultipartUploadHandler(filehash string, filename string, filesize int64) {
	//解析用户请求参数
	//r.ParseForm()
	//filehash := r.Form.Get("filehash")
	//filename := r.Form.Get("filename")
	//filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	//if err != nil {
	//	return
	//}

	//获得redis连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()

	//生成分块上传的初始化信息
	upinfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileName:   filename,
		FileSize:   filesize,
		UploadID:   fmt.Sprint("%x", time.Now().UnixNano()),
		ChunkSize:  64 * 1024 * 1024,
		ChunkCount: int(math.Ceil(float64(filesize) / (64 * 1024 * 1024))),
	}

	//将初始信息写入redis缓存
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "chunkcount", upinfo.ChunkCount)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filehash", upinfo.FileHash)
	rConn.Do("HSET", "MP_"+upinfo.UploadID, "filesize", upinfo.FileSize)

	//文件切分
	file, err := os.Open("/tmp/" + upinfo.FileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 定义每个切分文件的大小（字节数）
	chunkSize := 64 * 1024 * 1024

	// 读取并切分文件
	buffer := make([]byte, chunkSize)
	for i := 0; i < upinfo.ChunkCount; i++ {
		// 读取文件块
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if n == 0 {
			break
		}

		// 创建切分文件
		chunkFileName := fmt.Sprintf(upinfo.FileName+"%d", i)
		chunkFilePath := "/tmp/"
		chunkFile, err := os.Create(chunkFilePath + chunkFileName)
		if err != nil {
			log.Fatal(err)
		}
		defer chunkFile.Close()
	}
	//从配置文件中读取dataNode ip列表，将文件的分块按顺序分发到节点，更新redis数据
	for i := 0; i < upinfo.ChunkCount; i++ {
		IPAddress := utils.ReadFromUtil()
		tmpChunk := "/tmp/" + upinfo.FileName + strconv.Itoa(i)
		chunkName, err := os.Open(tmpChunk)
		if err != nil {
			log.Fatal(err)
		}
		for j := 0; j < 3; j++ {
			count := 0
			for k := j; k < len(IPAddress); k++ {
				conn, err := net.Dial("tcp", IPAddress[k]+":22")
				if err != nil {
					log.Fatal(err)
				}
				defer conn.Close()
				// 块分发
				fpath := "/data/" + upinfo.UploadID + "chkidx_" + strconv.Itoa(i) + "replica_" + strconv.Itoa(j)
				os.MkdirAll(path.Dir(fpath), 0777)
				fd, err := os.Create(fpath)
				if err != nil {
					return
				}
				defer fd.Close()

				buf := make([]byte, 1024*1024)
				for {
					// 读取文件块
					n, err := chunkName.Read(buf)
					if err != nil && err != io.EOF {
						log.Fatal(err)
					}

					// 发送文件块到服务器
					_, err = conn.Write(buffer[:n])
					if err != nil {
						log.Fatal(err)
					}
				}
				rConn.Do("HSET", "MP_"+upinfo.UploadID, "chkidx_"+strconv.Itoa(i)+"replica_"+strconv.Itoa(j), IPAddress[k])

				count++
				if k == j-1 && count < len(IPAddress) {
					k = 0
				}
			}
		}
	}

	//响应初始化数据返回客户端
}

// 上传文件分块
//func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
//	//解析用户请求参数
//	r.ParseForm()
//	uploadID := r.Form.Get("uploadid")
//	chunkIndex := r.Form.Get("index")
//
//	//获得redis连接池中的连接
//	rConn := rPool.RedisPool().Get()
//	defer rConn.Close()
//
//	//获得文件句柄， 用于存储分块内容
//	fpath := "/data/" + uploadID + "/" + chunkIndex
//	os.MkdirAll(path.Dir(fpath), 0777)
//	fd, err := os.Create(fpath)
//	if err != nil {
//		w.WriteHeader(401)
//		return
//	}
//	defer fd.Close()
//
//	buf := make([]byte, 1024*1024)
//	for {
//		n, err := r.Body.Read(buf)
//		fd.Write(buf[:n])
//		if err != nil {
//			break
//		}
//	}
//
//	//更新redis缓存状态
//	rConn.Do("HSET", "MP_"+uploadID, "chkidx_"+chunkIndex, 1)
//	//返回处理结果
//	io.WriteString(w, "Success")
//}

// 通知上传合并
//func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
//	//解析请求参数
//	r.ParseForm()
//	upid := r.Form.Get("uploadid")
//	//filehash := r.Form.Get("filehash")
//	//filesize := r.Form.Get("filesize")
//	//filename := r.Form.Get("filename")
//
//	//获得redis连接池中的连接
//	rConn := rPool.RedisPool().Get()
//	defer rConn.Close()
//
//	//通过uploadid查询redis并判断所有分块是否上传完成
//	data, err := redis.Values(rConn.Do("HGETALL", "MP_"+upid))
//	if err != nil {
//		return
//	}
//	totalCount := 0
//	chunkCount := 0
//
//	for i := 0; i < len(data); i += 2 {
//		k := string(data[i].([]byte))
//		v := string(data[i+1].([]byte))
//		if k == "chunkcount" {
//			totalCount, _ = strconv.Atoi(v)
//		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
//			chunkCount += 1
//		}
//	}
//
//	if totalCount != chunkCount {
//		return
//	}
//
//	//合并分块
//
//	//更新唯一文件表
//
//	//响应处理结果
//
//
