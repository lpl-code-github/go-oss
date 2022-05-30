package utils

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"runtime"
	"strconv"
	"strings"
)

func GetOffsetFromHeader(h http.Header) int64 {
	// 从header中获取range
	byteRange := h.Get("range")
	if len(byteRange) < 7 {
		return 0
	}
	if byteRange[:6] != "bytes=" {
		return 0
	}

	// 将bytes=<first>- 中的<first>-切取
	bytePos := strings.Split(byteRange[6:], "-")
	// 将字符串转为int64返回
	offset, _ := strconv.ParseInt(bytePos[0], 0, 64)
	return offset
}

func GetHashFromHeader(h http.Header) string {
	// 从header中获取digest
	digest := h.Get("digest")
	// 如果digest < 9 返回空字符串
	if len(digest) < 9 {
		return ""
	}
	// 如果digest 前8为不为SHA-256= 返回空字符串
	if digest[:8] != "SHA-256=" {
		return ""
	}

	// 否则返回8位以后，就是对象散列值的Base64编码
	return digest[8:]
}

func GetSizeFromHeader(h http.Header) int64 {
	// 从header中获取content-length，并将一个字符串转换一个数字
	size, _ := strconv.ParseInt(h.Get("content-length"), 0, 64)
	return size
}

func CalculateHash(r io.Reader) string {
	// 调用sha256.New生成变量h
	h := sha256.New()
	// 从r中读取数据并写入h
	io.Copy(h, r)

	// h对写入的数据通过h.Sum计算其hash值的二进制数据，通过base64.StdEncoding.EncodeToString进行编码
	return base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(h.Sum(nil))))
}
func PrintCallerName() string {
	pc, _, _, _ := runtime.Caller(2)
	return runtime.FuncForPC(pc).Name()
}
