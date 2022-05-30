package bucket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
	"strconv"
	"strings"
)

// 分页结构体
type bucketInfo struct {
	Size int64    `json:"size"` // 数据总长度
	Data []string `json:"data"` // 每页的数据
}

const SIZE = 4 // 每页显示条数

func get(w http.ResponseWriter, r *http.Request) {
	// 从路径中获得分页参数
	index, _ := strconv.Atoi(strings.Split(r.URL.EscapedPath(), "/")[2])
	// 获得桶名
	bucket := r.Header.Get("bucket")

	mapping := es.GetAllMapping()

	if bucket != "" { // 桶名不为空 则是查询单个
		myLog.Info.Println(fmt.Sprintf("查询桶 %s", bucket))
		unescape, _ := url.QueryUnescape(bucket)
		var result = make([]string, 0)
		for _, m := range mapping {
			if strVagueQuery(unescape, m) {
				result = append(result, m)
			}
		}

		helper := pageHelper(index, result) // 分页

		marshal, _ := json.Marshal(helper)
		w.WriteHeader(http.StatusOK)
		w.Write(marshal)
		return
	}

	// 否则是查询全部

	helper := pageHelper(index, mapping) // 分页
	marshal, _ := json.Marshal(helper)

	myLog.Info.Println(fmt.Sprintf("查询全部桶，第 %d 页", index))
	w.WriteHeader(http.StatusOK)
	w.Write(marshal)
}

// 分页
func pageHelper(page int, data []string) bucketInfo {
	size := len(data)
	info := bucketInfo{int64(size), nil}
	if size == 0 { //如果长度为0 直接返回
		info.Size = 0
		return info
	}

	metadata := make([]string, 0)
	start := (page - 1) * SIZE
	end := page * SIZE
	if start > len(data) {
		fmt.Println([]int{})
		return info
	}
	if len(data) < end {
		end = len(data)
	}

	metadata = data[start:end]
	info.Data = metadata

	return info
}

// 比较前一个字符串是否与后一个相同
func strVagueQuery(a string, b string) bool {
	return strings.Contains(b, a)
}
