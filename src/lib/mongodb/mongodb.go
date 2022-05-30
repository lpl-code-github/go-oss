package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

var uri = os.Getenv("MONGO_SERVER")

var ctx = context.Background()

// 初始化的客户端
var client *mongo.Client

func init() {
	// 设置客户端连接配置
	clientOptions := options.Client().ApplyURI(uri)
	// 连接到MongoDB
	var err error
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	// 检查连接
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// InsertOperation 插入操作记录
func InsertOperation(operation string) {
	date := time.Now().Format("2006-01-02")
	info := OperationInfo{operation, time.Now().Format("15:04:05"), date}
	insertOperationInfo(info)
	//通过date去更新查找所有的OperationInfo
	operationInfo := findOperationInfo(date)

	if len(operationInfo) == 1 { // 如果就一条数据 插入
		insertOperation(date, operationInfo)
	} else { // 否则更新
		update(date, operationInfo)
	}

}

type OperationInfo struct {
	Operation string
	Time      string
	Date      string
}

type Operation struct {
	Date string
	Data []OperationInfo
}

// 更新当日所有记录
func update(date string, operationInfo []*OperationInfo) int64 {
	c := client.Database("system").Collection("operation")
	update := bson.D{{"$set", bson.D{{"data", operationInfo}}}}
	ur, err := c.UpdateMany(ctx, bson.D{{"date", date}}, update)
	if err != nil {
		log.Fatal(err)
	}
	return ur.ModifiedCount
}

// 插入当日所有操作记录表
func insertOperation(date string, operationInfo []*OperationInfo) interface{} {
	collection := client.Database("system").Collection("operation")
	insertResult, err := collection.InsertOne(ctx, bson.D{{"date", date}, {"data", operationInfo}})
	if err != nil {
		log.Fatal(err)
	}
	return insertResult.InsertedID
}

// 插入当日单个操作记录
func insertOperationInfo(o OperationInfo) interface{} {
	collection := client.Database("system").Collection("operationInfo")
	insertResult, err := collection.InsertOne(ctx, o)
	if err != nil {
		log.Fatal(err)
	}
	return insertResult.InsertedID
}

// 查找当天多个操作记录
func findOperationInfo(date string) []*OperationInfo {
	// 定义一个数组用来存储查询结果
	var results []*OperationInfo
	collection := client.Database("system").Collection("operationInfo")
	// 把bson.D{{}}作为一个filter来匹配所有文档
	cur, err := collection.Find(ctx, bson.D{{"date", date}})
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(ctx) {
		var info OperationInfo
		err := cur.Decode(&info)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &info)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// 完成后关闭游标
	err = cur.Close(ctx)
	if err != nil {
		panic(err)
	}

	return results
}

// FindOperationAll 查找所有的记录
func FindOperationAll(index int64) (int64, []*Operation) {
	// 将分页参数传递给Find()
	findOptions := options.Find()
	findOptions.SetSkip(5 * index)
	findOptions.SetLimit(5) //显示行数
	findOptions.SetSort(bson.D{{"date", -1}})
	// 定义一个数组用来存储查询结果
	var results []*Operation
	collection := client.Database("system").Collection("operation")
	// 把bson.D{{}}作为一个filter来匹配所有文档
	cur, err := collection.Find(ctx, bson.D{}, findOptions)

	if err != nil {
		log.Fatal(err)
	}
	documents, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	// 查找多个文档返回一个光标
	// 遍历游标允许我们一次解码一个文档
	for cur.Next(ctx) {
		// 创建一个值，将单个文档解码为该值
		var operation Operation
		err := cur.Decode(&operation)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, &operation)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// 完成后关闭游标
	err = cur.Close(ctx)
	if err != nil {
		panic(err)
	}

	return documents, results
}
