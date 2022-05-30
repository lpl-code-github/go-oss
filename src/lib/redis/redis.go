package redis

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"os"
	"oss/src/lib/myLog"
	"strconv"
	"strings"
	"time"
)

// 单机模式
//var ctx = context.Background()
//var rdb *redis.Client
//
//func init() {
//	ctx = context.Background()
//
//	rdb = redis.NewClient(&redis.Options{
//		Addr:     "127.0.0.1:6379",
//		Password: "Lpl0618.", // no password set
//		DB:       0,          // use default DB
//		PoolSize: 10,
//	})
//}

// 集群模式
var ctx = context.Background()
var rdb *redis.ClusterClient

func init() {
	redisClusterAddrsString := os.Getenv("REDIS_CLUSTER")
	redisClusterAddr := strings.Split(redisClusterAddrsString, ",")

	rdb = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    redisClusterAddr,
		Password: os.Getenv("REDIS_PASSWORD"),
		PoolSize: 20,
	})
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}

}

func RedisSet(key string, value interface{}) string {
	err := rdb.Set(ctx, key, value, 0).Err()
	if err != nil {
		myLog.Error.Println(err)
		panic(err)
	}
	return "ok"
}

func RedisGet(key string) string {
	value, err := rdb.Get(ctx, key).Result()
	if err != nil {
		myLog.Error.Println(err)
	}
	return value
}

func RedisDelete(key string) string {
	result, err := rdb.Del(ctx, key).Result()
	if err != nil {
		myLog.Error.Println(err)
	}
	return strconv.FormatInt(result, 10)
}

func RedisClusterMget(keys []string) map[string]int64 {
	result := make(map[string]int64)
	pipeline := rdb.Pipeline()

	cmds, err := pipeline.Pipelined(ctx, func(pipe redis.Pipeliner) error {
		for _, key := range keys {
			pipeline.Get(ctx, key)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	for i, cmd := range cmds {
		num, _ := strconv.Atoi(cmd.(*redis.StringCmd).Val())
		result[keys[i]] = int64(num)
	}
	return result
}

// RedisGetKeys 通配符获取key
func RedisGetKeys(keys string) []string {
	//log.Printf("通配符查找包含%s的key", keys)
	result := make([]string, 0)
	err := rdb.ForEachSlave(ctx, func(ctx context.Context, rdb *redis.Client) error {
		iter := rdb.Scan(ctx, 0, keys, 0).Iterator()
		for iter.Next(ctx) {
			key := iter.Val()
			result = append(result, key)
		}
		return iter.Err()
	})
	if err != nil {
		panic(err)
	}

	return result
}

func RedisIncrAndEx(key string) string {
	//过期日期为明年1月1日
	s, _ := strconv.Atoi(strings.Split(key, "-")[0])
	expirationDate := fmt.Sprintf("%d-01-01", s+1)
	// 计算相差天数
	expirationDayTime := getTimeArr(key, expirationDate)
	// 过期时间
	expiration := time.Duration(expirationDayTime) * time.Second
	// setnx如果存在则不设置
	rdb.SetNX(ctx, key, 0, expiration)
	// 自增操作
	result, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		myLog.Error.Println(err)
		panic(err)
	}
	return strconv.FormatInt(result, 10)
}

func RedisIncr(key string) string {
	// setnx如果存在则不设置
	rdb.SetNX(ctx, key, 0, 0)
	result, err := rdb.Incr(ctx, key).Result()
	if err != nil {
		myLog.Error.Println(err)
		panic(err)
	}
	return strconv.FormatInt(result, 10)
}

func getTimeArr(start, end string) int64 {
	timeLayout := "2006-01-02"
	loc, _ := time.LoadLocation("Local")
	// 转成时间戳
	startUnix, _ := time.ParseInLocation(timeLayout, start, loc)
	endUnix, _ := time.ParseInLocation(timeLayout, end, loc)
	startTime := startUnix.Unix()
	endTime := endUnix.Unix()
	// 相差秒数
	seconds := endTime - startTime
	return seconds
}
