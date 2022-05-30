package objects

import (
	"fmt"
	"net/http"
	"net/url"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
	"oss/src/lib/redis"
	"oss/src/lib/utils"
	"strings"
	"time"
)

func put(w http.ResponseWriter, r *http.Request) {
	// 从请求头中获取对象hash值
	hash := utils.GetHashFromHeader(r.Header)
	if hash == "" {
		myLog.Error.Println("请求头中缺少hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// 获得桶名
	bucket := r.Header.Get("bucket")
	if bucket == "" {
		myLog.Error.Println("请求头中缺少桶名")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 计算hash值的长度
	size := utils.GetSizeFromHeader(r.Header)
	// storeObject存储对象
	c, e := storeObject(r.Body, hash, size)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(c)
		return
	}
	if c != http.StatusOK {
		w.WriteHeader(c)
		return
	}

	// 获取对象名
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	//unescape, _ := url.QueryUnescape(name)
	// 新增版本
	e = es.AddVersion(bucket, name, hash, size)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
	}
	unescape, _ := url.QueryUnescape(name)
	myLog.Info.Println(fmt.Sprintf("上传对象 %s", unescape))
	// 上传成功,记录请求的时间和次数,日历图显示
	redis.RedisIncrAndEx(time.Now().Format("2006-01-02"))
}
