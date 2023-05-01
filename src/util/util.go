package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"hash"
	"io"
	"os"
	"path/filepath"
)

type Sha1Stream struct {
	_sha1 hash.Hash
}

func (obj *Sha1Stream) Update(data []byte) {
	if obj._sha1 == nil {
		obj._sha1 = sha1.New()
	}
	obj._sha1.Write(data)
}

func (obj *Sha1Stream) Sum() string {
	return hex.EncodeToString(obj._sha1.Sum([]byte("")))
}

func Sha1(data []byte) string {
	_sha1 := sha1.New()
	_sha1.Write(data)
	return hex.EncodeToString(_sha1.Sum([]byte("")))
}

func FileSha1(file io.Reader) string {
	var err error
	_sha1 := sha1.New()
	if _, err = io.Copy(_sha1, file); err != nil {
		return err.Error()
	}

	return hex.EncodeToString(_sha1.Sum(nil))
}

func MD5(data []byte) string {
	_md5 := md5.New()
	_md5.Write(data)

	return hex.EncodeToString(_md5.Sum([]byte("")))
}

func FileMD5(file *os.File) string {
	_md5 := md5.New()
	io.Copy(_md5, file)

	return hex.EncodeToString(_md5.Sum(nil))
}

// 路径是否存在
func PathExists(path string) (bool, error) {
	var err error
	if _, err := os.Stat(path); err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

// 获取文件大小
func GetFileSize(filename string) (bit int64, err error) {
	err = filepath.Walk(filename, func(path string, info os.FileInfo, err error) error {
		bit = info.Size()
		return nil
	})
	return
}

// 判定文件是否存在
func IsFileExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}
