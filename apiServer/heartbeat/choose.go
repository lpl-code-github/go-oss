package heartbeat

import (
	"math/rand"
)

// ChooseRandomDataServers 能够返回多个随机数据服务节点
// 参数n：需要多少个随机数据服务节点，参数exclude 返回随机数据服务节点不能包含哪些
func ChooseRandomDataServers(n int, exclude map[int]string) (ds []string) {
	// 候选节点数组
	candidates := make([]string, 0)

	// exclude键值调换为新map
	reverseExcludeMap := make(map[string]int)
	for id, addr := range exclude {
		reverseExcludeMap[addr] = id
	}

	// 获得现在所有保持心跳的数据服务节点
	servers := GetDataServers()
	for i := range servers {
		// 服务节点地址s
		s := servers[i]
		// 是否为排除地址？
		_, excluded := reverseExcludeMap[s]
		// 不是排除地址
		if !excluded {
			// 将服务节点添加到候选节点
			candidates = append(candidates, s)
		}
	}

	length := len(candidates)
	// 候选节点小于需要的节点，无法满足，直接返回
	if length < n {
		return
	}

	// 如果候选节点大于等于需要的随机数据服务节点数
	// 将0~length-1的所有整数乱序排列返回一个数组
	p := rand.Perm(length)

	for i := 0; i < n; i++ {
		// 取前n个作为candidates数组的下标取数据节点地址 返回ds数组
		ds = append(ds, candidates[p[i]])
	}
	return
}
