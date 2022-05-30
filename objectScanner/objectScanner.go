package main

import (
	"fmt"
	"log"
	"os"
	"oss/apiServer/heartbeat"
	"oss/apiServer/objects"
	"oss/src/lib/es"
	"oss/src/lib/mongodb"
	"oss/src/lib/myLog"
	"oss/src/lib/redis"
	"oss/src/lib/utils"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	servers := heartbeat.GetDataServers()

	// 选取一个节点，如果已经节点存在（目录存在），获取objects目录下的所有文件，执行检查
	// 如果所有不存在，重新尝试
	for i := 0; i < len(servers); i++ {
		// 获取objects目录下的所有文件
		files, e := filepath.Glob(os.Getenv("STORAGE_ROOT") + strconv.Itoa(i) + "/objects/*")
		if e != nil {
			// 如果是最后一次循环
			if i == len(servers)-1 {
				myLog.Error.Println(e)
				return
			}
			continue
		}
		// 遍历这些文件
		for j := range files {
			// 从文件中获得对象的散列值
			hash := strings.Split(filepath.Base(files[j]), ".")[0]
			// 调用verify检查
			verify(hash)
		}
	}

	myLog.Trace.Println("全盘数据扫描修复")
	// 记录总请求次数
	redis.RedisIncr("upholdNum")
	mongodb.InsertOperation(fmt.Sprintf("进行了对象数据全盘扫描修复的操作"))
}

func verify(hash string) {
	log.Println("verify", hash)
	// 通过hash值获取该hash值对应的大小
	var size int64 = 0
	mapping := es.GetAllMapping()
	for _, m := range mapping {
		s, e := es.SearchHashSize(m, hash)
		if e != nil {
			myLog.Error.Println(e)
			return
		}
		if s != 0 {
			size = s
			break
		}
	}

	// 以对象的hash值和大小  创建一个对象数据流
	stream, e := objects.GetStream(hash, size)
	if e != nil {
		myLog.Error.Println(e)
		return
	}

	// 计算对象的散列值
	d := utils.CalculateHash(stream)
	if d != hash {
		myLog.Error.Println(fmt.Sprintf("对象哈希不匹配，计算=%s，请求=%s", d, hash))
	}
	// 关闭数据对象流
	stream.Close()
}
