package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

func main() {

	fileName := "googlechrome.dmg"
	//文件切分
	file, err := os.Open("/tmp/" + fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 定义每个切分文件的大小（字节数）
	chunkSize := 64 * 1024 * 1024

	// 读取并切分文件
	buffer := make([]byte, chunkSize)
	for i := 0; i < 4; i++ {
		// 读取文件块
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		if n == 0 {
			break
		}

		// 创建切分文件
		chunkFileName := fmt.Sprintf(fileName+"%d", i)
		chunkFilePath := "/tmp/"
		chunkFile, err := os.Create(chunkFilePath + chunkFileName)
		if err != nil {
			log.Fatal(err)
		}
		defer chunkFile.Close()
	}
}
