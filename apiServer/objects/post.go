package objects

import (
	"net/http"
	"net/url"
	"oss/apiServer/heartbeat"
	"oss/apiServer/locate"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
	"oss/src/lib/rs"
	"oss/src/lib/utils"
	"strconv"
	"strings"
)

func post(w http.ResponseWriter, r *http.Request) {
	// 获得对象的名字
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	//unescape, _ := url.QueryUnescape(name)
	// 从请求头获得对象size
	size, e := strconv.ParseInt(r.Header.Get("size"), 0, 64)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	// 获得桶名
	bucket := r.Header.Get("bucket")
	if bucket == "" {
		myLog.Error.Println("请求头中缺少桶名")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 从对象中获得hash值
	hash := utils.GetHashFromHeader(r.Header)
	if hash == "" {
		myLog.Error.Println("请求头中缺少hash")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// 对hash值定位，如果该hash值已经存在，往元数据服务添加新版本
	if locate.Exist(url.PathEscape(hash)) {
		e = es.AddVersion(bucket, name, hash, size)
		if e != nil {
			myLog.Error.Println(e)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		return
	}

	// 如果hash值不存在，随机选出6个服务节点
	ds := heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS, nil)
	if len(ds) != rs.ALL_SHARDS {
		myLog.Error.Println("找不到足够的数据服务节点")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	// 生成数据流
	stream, e := rs.NewRSResumablePutStream(ds, name, url.PathEscape(hash), size)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	myLog.Info.Println("获取分片上传token")
	// 调用ToToken方法生成一个token字符串 放入相应头部
	w.Header().Set("location", "/temp/"+url.PathEscape(stream.ToToken()))
	// 返回201 已创建
	w.WriteHeader(http.StatusCreated)
}
