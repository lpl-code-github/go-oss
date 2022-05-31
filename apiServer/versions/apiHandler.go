package versions

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
	"strconv"
	"strings"
)

var limit = 5

type allObjInfo struct {
	Size int64
	Data []es.Metadata
}

func ApiHandler(w http.ResponseWriter, r *http.Request) {
	// 请求方法
	m := r.Method
	if m != http.MethodGet { // 如果不是get方法
		// 返回405方法错误
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 获得桶名
	bucket := strings.Split(r.URL.EscapedPath(), "/")[2]
	if bucket == "" {
		myLog.Error.Println("路径参数中缺少桶名")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 从路径中获得对象名
	//param := strings.Split(r.URL.EscapedPath(), "/")[2]
	name := strings.Split(r.URL.EscapedPath(), "/")[3]

	// 获得page
	pageIndex := r.URL.Query()["page"]
	page := 1
	var e error
	if len(pageIndex) != 0 {
		page, e = strconv.Atoi(pageIndex[0])
		if e != nil {
			log.Println(e)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	// 如果有参数 调用es包的SearchApiVersions，返回全部对象的元数据的数组
	result, e := GetAll(bucket, name)
	if e != nil {
		log.Println(e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 判断长度 如果长度=0直接返回
	size := len(result)
	info := allObjInfo{int64(size), nil}
	if size == 0 {
		info.Size = 0
		b, _ := json.Marshal(info)
		w.Write(b)
		return
	}

	// 如果长度不为0 将数据分页后返回
	info = pageHelper(page, result)
	myLog.Info.Println(fmt.Sprintf("查询桶：%s 全部对象的最新版本", bucket))
	b, _ := json.Marshal(info)
	w.Write(b)
}

// GetAll 获取所有对象最新版本
func GetAll(bucket string, name string) ([]es.Metadata, error) {
	from := 0
	size := 1000
	result := make([]es.Metadata, 0)
	//无限循环
	for {
		// 通过对象名 调用es包的SearchAllVersions，返回某个对象的元数据的数组
		metas, e := es.SearchApiVersions(bucket, name, from, size)
		// 如果报错
		if e != nil {
			// 打印错误并返回500
			return result, e
		}
		// 遍历某个对象元数据数组
		for i := range metas {
			result = append(result, metas[i])
		}
		// 如果长度数据长度不等于1000，此时元数据服务中没有更多的数据了
		if len(metas) != size {
			// 结束循环
			goto breakHere

		}
		//否则把from的值+1000进行下一次迭代
		from += size

	}
breakHere:
	return result, nil
}

func pageHelper(page int, data []es.Metadata) allObjInfo {
	size := len(data)
	info := allObjInfo{int64(size), nil}
	if size == 0 {
		info.Size = 0
		return info
	}
	metadata := make([]es.Metadata, 0)
	//手写分页
	start := (page - 1) * limit
	end := page * limit
	if start > len(data) {
		return info
	}
	if len(data) < end {
		end = len(data)
	}
	metadata = data[start:end]
	info.Data = metadata

	return info
}
