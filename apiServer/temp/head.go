package temp

import (
	"fmt"
	"net/http"
	"oss/src/lib/myLog"
	"oss/src/lib/rs"
	"strings"
)

func head(w http.ResponseWriter, r *http.Request) {
	// 用token恢复出stream
	token := strings.Split(r.URL.EscapedPath(), "/")[3]
	stream, e := rs.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 获取当前大小
	current := stream.CurrentSize()
	if current == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// 将当前大小放在响应头content-length中
	w.Header().Set("content-length", fmt.Sprintf("%d", current))
}
