package temp

import (
	"net/http"
	"os"
	"strings"
)

func del(w http.ResponseWriter, r *http.Request) {
	// 从请求头获取uuid
	uuid := strings.Split(r.URL.EscapedPath(), "/")[2]

	// 获取临时对象的 信息文件和数据文件
	infoFile := os.Getenv("STORAGE_ROOT") + "/temp/" + uuid
	datFile := infoFile + ".dat"

	// 删除临时对象信息文件和数据文件
	os.Remove(infoFile)
	os.Remove(datFile)
}
