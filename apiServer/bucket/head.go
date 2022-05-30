package bucket

import (
	"fmt"
	"net/http"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
)

func head(w http.ResponseWriter, r *http.Request) {
	// 获得桶名
	bucket := r.Header.Get("bucket")
	if bucket == "" {
		myLog.Error.Println("请求头中缺少桶名")
		//log.Println("missing object bucket in header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// 查找隐射
	httpCode := es.SearchMapping(bucket)

	myLog.Info.Println(fmt.Sprintf("查询桶 %s 是否存在", bucket))
	w.WriteHeader(httpCode)
}
