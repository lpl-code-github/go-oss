package checktool

import (
	"fmt"
	"net/http"
	"oss/src/lib/es"
	"oss/src/lib/mongodb"
	"oss/src/lib/myLog"
	"oss/src/lib/redis"
	"strconv"
	"strings"
)

func DeleteOldMetadata(w http.ResponseWriter, r *http.Request) {
	// 保留几个版本
	version := strings.Split(r.URL.EscapedPath(), "/")[2]

	// 请求方法
	m := r.Method
	if m != http.MethodGet { // 如果不是get方法
		// 返回405方法错误
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	mapping := es.GetAllMapping()
	for _, s := range mapping {
		// 搜索元数据服务中所有版本>=version+1的对象，保存在Bucket结构体的数组buckets中
		atoi, e := strconv.Atoi(version)
		if e != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		buckets, e := es.SearchVersionStatus(s, atoi+1)
		if e != nil {
			myLog.Error.Println(e)
			return
		}
		// 遍历buckets
		for i := range buckets {
			bucket := buckets[i]
			// 从该对象当前最小的版本号开始删除，直到最后还剩version+1个
			for v := 0; v < bucket.Doc_count-atoi; v++ {
				es.DelMetadata(s, bucket.Key, v+int(bucket.Min_version.Value))
			}
		}
	}

	myLog.Trace.Println(fmt.Sprintf("保留全部对象的%s个版本操作", version))
	// 记录总请求次数
	redis.RedisIncr("upholdNum")
	mongodb.InsertOperation(fmt.Sprintf("进行了保留对象版本操作：保留了全部对象的%s个版本", version))
	w.WriteHeader(http.StatusOK)
}
