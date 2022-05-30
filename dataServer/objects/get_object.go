package objects

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/url"
	"os"
	"oss/dataServer/locate"
	"oss/src/lib/myLog"
	"path/filepath"
	"strings"
)

func getFile(name string) string {
	// 找所有以<hash>.X开头的文件
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/" + name + ".*")
	if len(files) != 1 {
		return "" // 找不到返回空字符串
	}

	// 找到后计算其散列值
	file := files[0]
	h := sha256.New()
	sendFile(h, file)
	d := url.PathEscape(base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(h.Sum(nil)))))
	hash := strings.Split(file, ".")[2]

	// 如果计算的散列值和文件名中hash值不匹配则删除该对象 并返回空字符串
	if d != hash {
		myLog.Error.Println("对象哈希不匹配，删除", file)
		locate.Del(hash)
		os.Remove(file)
		return ""
	}
	return file
}
