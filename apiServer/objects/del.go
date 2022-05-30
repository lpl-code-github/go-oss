package objects

import (
	"fmt"
	"net/http"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
	"strings"
)

func del(w http.ResponseWriter, r *http.Request) {
	// 获取对象名称
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	//unescape, _ := url.QueryUnescape(name)
	// 获得桶名
	bucket := r.Header.Get("bucket")
	if bucket == "" {
		myLog.Error.Println("请求头中缺少桶名")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 调用SearchLatestVersion，通过对象名搜索该对象最新版本
	version, e := es.SearchLatestVersion(bucket, name)
	// 如果报错
	if e != nil {
		// 打印错误并返回500
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 插入一条新的元数据，将版本+1，size设置为0，hash为空字符串
	e = es.PutMetadata(bucket, name, version.Version+1, 0, "")
	// 如果报错
	if e != nil {
		// 打印错误并返回500
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	myLog.Info.Println(fmt.Sprintf("删除对象 %s", name))
}
