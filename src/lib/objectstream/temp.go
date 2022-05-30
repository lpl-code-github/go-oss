package objectstream

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

// TempPutStream 结构体 包含Server和Uuid
type TempPutStream struct {
	Server string
	Uuid   string
}

func NewTempPutStream(server, object string, size int64) (*TempPutStream, error) {
	// 发送post请求，访问数据服务的temp接口从而获得uuid
	request, e := http.NewRequest("POST", "http://"+server+"/temp/"+object, nil)
	if e != nil {
		return nil, e
	}

	request.Header.Set("size", fmt.Sprintf("%d", size))
	client := http.Client{}

	// 向数据服务发送请求
	response, e := client.Do(request)
	if e != nil {
		return nil, e
	}

	// 获取响应body，得到uuid
	uuid, e := ioutil.ReadAll(response.Body)
	if e != nil {
		return nil, e
	}

	// 返回TempPutStream，包含server和uuid
	return &TempPutStream{server, string(uuid)}, nil
}

func (w *TempPutStream) Write(p []byte) (n int, err error) {
	// 根据server uuid 以PATCH方法访问数据服务的temp接口，将需要写入的数据上传
	request, e := http.NewRequest("PATCH", "http://"+w.Server+"/temp/"+w.Uuid, strings.NewReader(string(p)))
	if e != nil {
		return 0, e
	}
	client := http.Client{}
	r, e := client.Do(request)
	if e != nil {
		return 0, e
	}

	if r.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("dataServer return http code %d", r.StatusCode)
	}

	return len(p), nil
}

func (w *TempPutStream) Commit(good bool) {
	// 默认delete访问数据服务层temp接口
	method := "DELETE"
	// 如果good为true，就put，访问数据层temp接口
	if good {
		method = "PUT"
	}
	request, _ := http.NewRequest(method, "http://"+w.Server+"/temp/"+w.Uuid, nil)
	client := http.Client{}
	client.Do(request)
}

func NewTempGetStream(server, uuid string) (*GetStream, error) {
	return newGetStream("http://" + server + "/temp/" + uuid)
}
