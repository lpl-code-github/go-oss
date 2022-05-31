package temp

import (
	"io"
	"net/http"
	"net/url"
	"oss/apiServer/locate"
	"oss/src/lib/es"
	"oss/src/lib/myLog"
	"oss/src/lib/rs"
	"oss/src/lib/utils"
	"strings"
)

func put(w http.ResponseWriter, r *http.Request) {
	// 获得桶名
	bucket := strings.Split(r.URL.EscapedPath(), "/")[2]
	if bucket == "" {
		myLog.Error.Println("路径参数中缺少桶名")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 从url中获取token
	token := strings.Split(r.URL.EscapedPath(), "/")[3]
	// 根据token 创建RSResumablePutStream结构体指针stream
	stream, e := rs.NewRSResumablePutStreamFromToken(token)
	if e != nil {
		myLog.Error.Println(e)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// 获得token当前大小
	current := stream.CurrentSize()
	//log.Printf("current=%d", current)
	if current == -1 {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// 获得offset偏移量
	offset := utils.GetOffsetFromHeader(r.Header)
	//log.Printf("url.offset=%d", offset)
	if current != offset { // 不一致 返回416
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}

	// 32000位的byte类型的数组
	bytes := make([]byte, rs.BLOCK_SIZE)
	// 读取http请求正文并写入stream中
	for {
		n, e := io.ReadFull(r.Body, bytes)
		if e != nil && e != io.EOF && e != io.ErrUnexpectedEOF {
			myLog.Error.Println(e)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		current += int64(n)
		if current > stream.Size { // 如果读到的总长度超出了对象的大小
			// 客户端上传的数据有误 删除临时对象
			stream.Commit(false)
			myLog.Error.Println("可恢复放置超出大小")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		// 如果某次读取的长度不到32000字节 且读取到总长度不等于对象的大小
		if n != rs.BLOCK_SIZE && current != stream.Size {
			// 本次客户端上传结束 还有后续数据需要上传
			return
		}

		stream.Write(bytes[:n])

		// 如果读取到的总长度等于对象的大小，说明客户端上传了对象的全部数据
		if current == stream.Size {
			// 将属于数据写入临时对象
			stream.Flush()
			// 生成一个临时对象读取流
			getStream, e := rs.NewRSResumableGetStream(stream.Servers, stream.Uuids, stream.Size)
			// 读取临时对象读取流中的数据并计算hash值
			hash := url.PathEscape(utils.CalculateHash(getStream))
			if hash != stream.Hash { // hash值不一致，说明客户端上传数据有误
				// 删除临时对象
				stream.Commit(false)
				myLog.Error.Println("可恢复的put已完成，但哈希不匹配")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// 检查散列值是否已经存在
			if locate.Exist(url.PathEscape(hash)) {
				// 存在 删除临时对象
				stream.Commit(false)
			} else {
				// 不存在，将临时对象转正
				stream.Commit(true)
			}

			// 添加对象新版本
			e = es.AddVersion(bucket, stream.Name, stream.Hash, stream.Size)
			if e != nil {
				myLog.Error.Println(e)
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}
	}
}
