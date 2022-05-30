package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"oss/apiServer/bucket"
	"oss/apiServer/checktool"
	"oss/apiServer/heartbeat"
	"oss/apiServer/locate"
	"oss/apiServer/logs"
	"oss/apiServer/objects"
	"oss/apiServer/system"
	"oss/apiServer/temp"
	"oss/apiServer/versions"
	"oss/src/lib/myLog"
	"time"
)

func main() {
	go heartbeat.ListenHeartbeat()
	go myLog.ListenLogExchange()
	go myLog.ReadLog(time.Now().Format("2006-01-02")) // 实时读取日志
	//go writeLog()                                     //测试方法
	http.HandleFunc("/bucket/", bucket.Handler)
	http.HandleFunc("/objects/", objects.Handler)
	http.HandleFunc("/temp/", temp.Handler)
	http.HandleFunc("/locate/", locate.Handler)
	http.HandleFunc("/versions/", versions.Handler)
	http.HandleFunc("/allVersions/", versions.ApiHandler)
	http.HandleFunc("/heartbeat", heartbeat.Handler)
	http.HandleFunc("/systemInfo/", system.Handler)
	http.HandleFunc("/nodeSystemInfo/", system.NodeSystemInfo)
	http.HandleFunc("/getLog/", logs.Handler)
	http.HandleFunc("/deleteOldMetadata/", checktool.DeleteOldMetadata)
	http.HandleFunc("/deleteOrphanServer", checktool.DeleteOrphan)
	http.HandleFunc("/objectScanner", checktool.ObjectScanner)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}

//测试
func writeLog() {
	var i = 0
	for {
		myLog.Trace.Println("I have something standard to say.")
		myLog.Info.Println(fmt.Sprintf("info myLog -%d", i))

		myLog.Warn.Println(fmt.Sprintf("warn myLog -%d", i))
		myLog.Error.Println("Something has failed")
		i++
		time.Sleep(2 * time.Second)
	}
}
