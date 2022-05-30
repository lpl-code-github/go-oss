package main

import (
	"fmt"
	"oss/src/lib/es"
	"oss/src/lib/mongodb"
	"oss/src/lib/myLog"
	"oss/src/lib/redis"
)

const MIN_VERSION_COUNT = 5

func main() {
	// 搜索元数据服务中所有版本>=6的对象，保存在Bucket结构体的数组buckets中
	mapping := es.GetAllMapping()
	for _, s := range mapping {
		buckets, e := es.SearchVersionStatus(s, MIN_VERSION_COUNT+1)
		if e != nil {
			myLog.Error.Println(e)
			return
		}

		// 遍历buckets
		for i := range buckets {
			bucket := buckets[i]
			// 从该对象当前最小的版本号开始删除，直到最后还剩5个
			for v := 0; v < bucket.Doc_count-MIN_VERSION_COUNT; v++ {
				es.DelMetadata(s, bucket.Key, v+int(bucket.Min_version.Value))
			}
		}
	}
	myLog.Trace.Println(fmt.Sprintf("保留全部对象的%d个版本操作", MIN_VERSION_COUNT))
	// 记录总请求次数
	redis.RedisIncr("upholdNum")
	mongodb.InsertOperation(fmt.Sprintf("进行了保留对象版本操作：保留了全部对象的%d个版本", MIN_VERSION_COUNT))
}
