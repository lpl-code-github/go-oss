package system

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"oss/apiServer/versions"
	"oss/src/lib/es"
	"oss/src/lib/mongodb"
	"oss/src/lib/myLog"
	"oss/src/lib/redis"
	"strconv"
	"strings"
	"time"
)

type Info struct {
	Obj       int64
	Put       int64
	Uphold    int64
	Echarts   map[string]int64
	Operation Operation
}
type Operation struct {
	OperationSize int64
	OperationData []*mongodb.Operation
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// r的成员Method变量记录了HTTP请求的方法
	m := r.Method

	// 检查请求方法
	if m != http.MethodGet { // 如果不是get请求，返回405
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 从路径中获得分页参数
	index, _ := strconv.Atoi(strings.Split(r.URL.EscapedPath(), "/")[2])

	all, operations := mongodb.FindOperationAll(int64(index))
	operation := Operation{all, operations}
	system := Info{
		Obj:       getObjNum(),
		Put:       getPutNum(),
		Uphold:    upholdNum(),
		Echarts:   getEcharts(),
		Operation: operation,
	}

	myLog.Info.Println("获取系统维护信息")
	b, _ := json.Marshal(system)
	w.Write(b)
}

func getObjNum() int64 {
	buckets := es.GetAllMapping()
	length := len(buckets)
	if length == 0 {
		return 0
	}
	var num int64 = 0
	for _, b := range buckets {
		metas, e := versions.GetAll(b, "")
		if e != nil {
			myLog.Error.Println(e)
		}
		num += int64(len(metas))
	}

	return num
}

func getPutNum() int64 {
	// 初始化结果
	var num int64 = 0
	putInfo := getEcharts()
	for p, _ := range putInfo {
		num += int64(putInfo[p])
	}
	return num
}
func upholdNum() int64 {
	upholdGet := redis.RedisGet("upholdNum")
	if upholdGet == "" {
		return 0
	}
	num, err := strconv.Atoi(upholdGet)
	if err != nil {
		log.Println(err)
	}
	var uphold = int64(num)
	return uphold
}

func getEcharts() map[string]int64 {
	result := make(map[string]int64)
	keys := redis.RedisGetKeys(fmt.Sprintf("%d*", time.Now().Year()))

	if len(keys) == 0 {
		return result
	}

	result = redis.RedisClusterMget(keys)
	return result
}
