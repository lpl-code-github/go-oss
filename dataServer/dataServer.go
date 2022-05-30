package main

import (
	"log"
	"net/http"
	"os"
	"oss/dataServer/heartbeat"
	"oss/dataServer/locate"
	"oss/dataServer/objects"
	"oss/dataServer/system"
	"oss/dataServer/temp"
)

func main() {
	locate.CollectObjects()
	go heartbeat.StartHeartbeat()
	go locate.StartLocate()
	http.HandleFunc("/systemInfo", system.Handler)
	http.HandleFunc("/objects/", objects.Handler)
	// 接口中包含/temp/中 执行temp.Handler
	http.HandleFunc("/temp/", temp.Handler)
	log.Fatal(http.ListenAndServe(os.Getenv("LISTEN_ADDRESS"), nil))
}
