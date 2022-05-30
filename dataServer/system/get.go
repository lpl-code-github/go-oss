package system

import (
	"encoding/json"
	"net/http"
	"oss/src/lib/myLog"
	"oss/src/lib/system"
)

func get(w http.ResponseWriter, r *http.Request) {
	info := system.GetInfo()

	marshal, err := json.Marshal(info)
	if err != nil {
		myLog.Error.Println(err)
	}
	w.Write(marshal)
}
