package main

import (
	"fmt"
	"log"
	"net/http"
	"oss/apiServer/heartbeat"
	"oss/src/lib/es"
	"oss/src/lib/mongodb"
	"oss/src/lib/myLog"
	"oss/src/lib/redis"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	mapping := es.GetAllMapping()
	servers := heartbeat.GetDataServers()

	for i := 0; i < len(servers); i++ {
		// 获取objects目录下的所有文件
		files, _ := filepath.Glob("/tmp/" + strconv.Itoa(i+1) + "/objects/*")

		// 遍历文件数组
		for j := range files {
			// 获取hash值
			hash := strings.Split(filepath.Base(files[j]), ".")[0]
			var hashInMetadata = false // 存在标记
			// 检查元数据服务中是否存在hash值
			for _, ma := range mapping {
				flag, e := es.HasHash(ma, hash)
				if e != nil {
					myLog.Error.Println(e)
					return
				}
				if flag == true {
					hashInMetadata = true
					break
				}
			}

			if !hashInMetadata {
				// 不存在删除
				del(servers[i], hash)
			}
		}
	}

	myLog.Trace.Println("删除无元数据引用的对象数据")
	// 记录总请求次数
	redis.RedisIncr("upholdNum")
	mongodb.InsertOperation(fmt.Sprintf("进行了删除无元数据引用的文件的操作"))
}

func del(addr string, hash string) {
	log.Println("delete", hash)
	url := "http://" + addr + "/objects/" + hash
	request, _ := http.NewRequest("DELETE", url, nil)
	client := http.Client{}
	client.Do(request)
}
