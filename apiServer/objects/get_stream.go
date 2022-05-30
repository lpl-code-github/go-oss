package objects

import (
	"fmt"
	"oss/apiServer/heartbeat"
	"oss/apiServer/locate"
	"oss/src/lib/rs"
)

// GetStream 获取文件流
func GetStream(hash string, size int64) (*rs.RSGetStream, error) {
	// 定位文件分片
	locateInfo := locate.Locate(hash)

	// 如果locateInfo长度小于4
	if len(locateInfo) < rs.DATA_SHARDS {
		// 定位失败
		return nil, fmt.Errorf("对象 %s 定位失败, result %v", hash, locateInfo)
	}

	// dataServers数组
	dataServers := make([]string, 0)

	// 如果locateInfo长度不等于6 说明有部分分片丢失
	if len(locateInfo) != rs.ALL_SHARDS {
		// 随机选取用于接收恢复分片的数据服务节点 以数组的形式保存在dataServers
		dataServers = heartbeat.ChooseRandomDataServers(rs.ALL_SHARDS-len(locateInfo), locateInfo)
	}

	return rs.NewRSGetStream(locateInfo, dataServers, hash, size)
}
