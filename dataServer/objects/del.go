package objects

import (
	"log"
	"net/http"
	"os"
	"oss/dataServer/locate"
	"path/filepath"
	"strings"
)

func del(w http.ResponseWriter, r *http.Request) {
	// 根据hash值搜索队列文件
	hash := strings.Split(r.URL.EscapedPath(), "/")[2]
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/objects/" + hash + ".*")
	if len(files) != 1 {
		return
	}
	log.Println(files)
	// 将该hash值移出对象定位缓存
	locate.Del(hash)

	// 将对象文件移动到garbage目录下
	os.Rename(files[0], os.Getenv("STORAGE_ROOT")+"/garbage/"+filepath.Base(files[0]))

}
