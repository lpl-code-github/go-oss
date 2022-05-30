package objects

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"oss/apiServer/locate"
	"oss/src/lib/utils"
)

func storeObject(r io.Reader, hash string, size int64) (int, error) {
	// 定位对象的散列值
	if locate.Exist(url.PathEscape(hash)) {
		// 如果存在 跳过后续上传操作直接返回200
		return http.StatusOK, nil
	}

	// 如果定位到不存在对象 实际进行存储：

	// putStream生成对象的写入流stream
	stream, e := putStream(url.PathEscape(hash), size)
	if e != nil {
		return http.StatusInternalServerError, e
	}

	// 对象写入stream同时当读取 ->r
	reader := io.TeeReader(r, stream)

	// 从reader中读取数据进行hash计算
	d := utils.CalculateHash(reader)
	// 比较对象的hash值和传入的hash值
	if d != hash {
		// 不一致 删除对象
		stream.Commit(false)
		return http.StatusBadRequest, fmt.Errorf("对象哈希不匹配，计算=%s，请求=%s", d, hash)
	}

	// 一致，将临时对象转正
	stream.Commit(true)
	return http.StatusOK, nil
}
