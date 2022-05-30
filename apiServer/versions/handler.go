package versions

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
	"strings"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	// 请求方法
	m := r.Method
	if m != http.MethodGet { // 如果不是get方法
		// 返回405方法错误
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 获得桶名
	bucket := r.Header.Get("bucket")
	if bucket == "" {
		myLog.Error.Println("请求头中缺少桶名")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 分页参数
	from := 0
	size := 1000
	result := make([]es.Metadata, 0)
	// 获取对象名称
	name := strings.Split(r.URL.EscapedPath(), "/")[2]
	//unescape, _ := url.QueryUnescape(name)
	//无限循环
	for {
		// 通过对象名 调用es包的SearchAllVersions，返回某个对象的元数据的数组
		metas, e := es.SearchAllVersions(bucket, name, from, size)
		// 如果报错
		if e != nil {
			// 打印错误并返回500
			myLog.Error.Println(e)
			w.WriteHeader(http.StatusInternalServerError)
			return
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
	// 序列化json
	b, _ := json.Marshal(result)
	// 写入响应体
	w.Write(b)
	unescape, _ := url.QueryUnescape(name)
	myLog.Info.Println(fmt.Sprintf("查询桶：%s 对象名：%s 的全部版本", bucket, unescape))
}
