package temp

import (
	"io"
	"net/http"
	"os"
	"oss/src/lib/myLog"
	"strings"
)

//get 函数打开$STORAGE_ROOT/temp/<uuid>.dat 文件并将其内容作为http的响应正文输出。
func get(w http.ResponseWriter, r *http.Request) {
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]
	f, e := os.Open(os.Getenv("STORAGE_ROOT") + "/temp/" + uuid + ".dat")
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	defer f.Close()
	io.Copy(w, f)
}
