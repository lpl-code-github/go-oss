package objects

import (
	"fmt"
	"oss/apiServer/heartbeat"
	"oss/src/lib/rs"
)

func putStream(hash string, size int64) (*rs.RSPutStream, error) {
	// 获取全部服务节点
	servers := heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS, nil)

	// 如果获取到的服务节点 不等于常量6
	if len(servers) != rs.ALL_SHARDS {
		return nil, fmt.Errorf("cannot find enough dataServer")
	}

	// 生成一个数据流
	return rs.NewRSPutStream(servers, hash, size)
}
