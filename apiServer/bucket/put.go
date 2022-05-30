package bucket

import (
	"fmt"
	"net/http"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
)

func put(w http.ResponseWriter, r *http.Request) {
	// 获得桶名
	bucket := r.Header.Get("bucket")
	if bucket == "" {
		myLog.Error.Println("请求头中缺少桶名")
		//log.Println("missing object bucket in header")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := es.AddMapping(bucket)
	if err != nil {
		myLog.Error.Println(err)
		//log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	myLog.Info.Println(fmt.Sprintf("添加桶 %s", bucket))
	w.WriteHeader(http.StatusCreated)
}
