package heartbeat

import (
	"encoding/json"
	"net/http"
)

// Handler 参数w：写入HTTP的响应，参数r：http请求对象
func Handler(w http.ResponseWriter, r *http.Request) {
	// r的成员Method变量记录了HTTP请求的方法
	m := r.Method

	// 检查请求方法
	if m != http.MethodGet { // 如果不是get请求，返回405
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// 获取数据节点
	servers := GetDataServersMap()

	heartbeatServers, _ := json.Marshal(servers)

	w.Write(heartbeatServers)
}
