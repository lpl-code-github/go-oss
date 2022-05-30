package temp

import (
	"fmt"
	"log"
	"net/http"
	"oss/src/lib/rs"
	"strings"
)

func head(w http.ResponseWriter, r *http.Request) {
	// 用token恢复出stream
	token := strings.Split(r.URL.EscapedPath(), "/")[2]
	stream, e := rs.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		log.Println(e)
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
