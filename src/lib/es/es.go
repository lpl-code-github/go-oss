package es

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	url2 "net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Metadata 元数据结构体
type Metadata struct {
	Name    string
	Version int
	Size    int64
	Hash    string
	Time    int64
}

// 搜索元数据的结构体
type searchResult struct {
	Hits struct {
		Total struct {
			Value    int
			Relation string
		}
		Hits []struct {
			Source Metadata `json:"_source"`
		}
	}
}

// Log log结构体
type Log struct {
	OsName   string `json:"osName"`
	Level    string `json:"level"`
	DateTime int64  `json:"dateTime"`
	Content  string `json:"content"`
}

// 搜索日志的响应结构体
type searchLogResult struct {
	Hits struct {
		Total struct {
			Value    int
			Relation string
		}
		Hits []struct {
			Source Log `json:"_source"`
		}
	}
}

// 搜索日志的请求结构体
type searchLogBody struct {
	Query struct {
		Bool struct {
			Must []interface{} `json:"must"`
		} `json:"bool"`
	} `json:"query"`
	Sort []map[string]string `json:"sort"`
	From int                 `json:"from"`
	Size int                 `json:"size"`
}

// es 连接响应体
type esRespResult struct {
	Name        string `json:"name"`
	ClusterName string `json:"cluster_name"`
	ClusterUuid string `json:"cluster_uuid"`
	Version     struct {
		Number                           string    `json:"number"`
		BuildFlavor                      string    `json:"build_flavor"`
		BuildType                        string    `json:"build_type"`
		BuildHash                        string    `json:"build_hash"`
		BuildDate                        time.Time `json:"build_date"`
		BuildSnapshot                    bool      `json:"build_snapshot"`
		LuceneVersion                    string    `json:"lucene_version"`
		MinimumWireCompatibilityVersion  string    `json:"minimum_wire_compatibility_version"`
		MinimumIndexCompatibilityVersion string    `json:"minimum_index_compatibility_version"`
	} `json:"version"`
	Tagline string `json:"tagline"`
}

var (
	available []string // 可用节点
	esAddr    = strings.Split(os.Getenv("ES_SERVER"), ",")
)

func init() {
	available = esAddr
	go apendAddr()
}

// 遍历所有节点
func apendAddr() {
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	for {
		result := make([]string, 0)
		for _, addr := range esAddr {
			url := fmt.Sprintf("http://%s", addr)

			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			r, _ := ioutil.ReadAll(resp.Body)
			esResp := &esRespResult{}
			err = json.Unmarshal(r, esResp)
			if err != nil {
				log.Println(err)
			}
			if esResp.Name != "" {
				result = append(result, addr)
			}
		}
		available = result
		// 延时5秒
		time.Sleep(5 * time.Second)
	}
}

// 选取一个节点
func choose(availableAddrs []string) string {
	// 从可用节点中随机选取一个节点
	addrIndex := rand.Intn(len(available))
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	url := fmt.Sprintf("http://%s", available[addrIndex])
	_, err := client.Get(url)
	if err != nil {
		return choose(availableAddrs)
	}
	return available[addrIndex]
}

// AddMapping 创建映射
func AddMapping(mapping string) error {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata_%s",
		choose(available), mapping)

	body := fmt.Sprintf(`
		{
    		"mappings": {
				"properties": {
            		"name": {
                		"type": "text",
                		"index": "true",
                		"fielddata":true,
                		"fields":{
                    		"subField":{
								"type":"keyword",
                        		"ignore_above":256
                    		}
                		}

            		},
            		"version": {
                		"type": "integer",
                		"index": "true"
            		},
            		"size": {
                		"type": "integer"
            		},
            		"hash": {
                		"type": "text",
                		"fields":{
                    		"subField":{
                        		"type":"keyword",
                        		"ignore_above":256
                    		}
                		}
            		}
        		}
    		}
	}
	`)
	request, _ := http.NewRequest("PUT", url, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	r, e := client.Do(request)
	if e != nil {
		return e
	}

	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to add mapping: %d", r.StatusCode)
		return e
	}
	return nil
}

// DeleteMapping 删除映射
func DeleteMapping(mapping string) error {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata_%s",
		choose(available), mapping)

	request, _ := http.NewRequest("DELETE", url, nil)
	request.Header.Set("Content-Type", "application/json")
	r, e := client.Do(request)
	if e != nil {
		return e
	}

	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to delete mapping: %d", r.StatusCode)
		return e
	}
	return nil
}

// GetAllMapping 查找全部映射
func GetAllMapping() []string {
	var bucketData []string

	url := fmt.Sprintf("http://%s/_mapping",
		choose(available))
	r, e := http.Get(url)
	if e != nil {
		log.Println(e)
	}
	if r.StatusCode != http.StatusOK {
		log.Println("fail to mapping")
	}
	mapping, _ := ioutil.ReadAll(r.Body)

	event := make(map[string]interface{})
	err := json.Unmarshal(mapping, &event)
	if err != nil {
		log.Println(err)
		return bucketData
	}

	// 解决map无序遍历的问题
	keys := make([]string, 0, len(event))
	for k := range event {
		if k == "log" {
			continue
		}
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		replace := strings.Replace(k, "metadata_", "", 1)
		bucketData = append(bucketData, replace)
	}

	return bucketData
}

// SearchMapping 搜索映射
func SearchMapping(name string) int {
	url := fmt.Sprintf("http://%s/metadata_%s//_mapping", choose(available), name)
	r, e := http.Get(url)
	if e != nil {
		log.Println(e)
	}
	return r.StatusCode
}

func getMetadata(mapping string, name string, versionId int) (meta Metadata, e error) {
	url := fmt.Sprintf("http://%s/metadata_%s/_doc/%s_%d/_source",
		choose(available), mapping, name, versionId)
	r, e := http.Get(url)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to get %s_%d: %d", name, versionId, r.StatusCode)
		return
	}
	result, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(result, &meta)
	return
}

func SearchLatestVersion(mapping string, name string) (meta Metadata, e error) {
	client := http.Client{}
	//url := fmt.Sprintf("http://%s/metadata/_search?q=name:%s&size=1&sort=version:desc",
	//	os.Getenv("ES_SERVER"),name)
	url := fmt.Sprintf("http://%s/metadata_%s/_search", choose(available), mapping)
	body := fmt.Sprintf(`
		{
  		  "query": {
            "match_phrase": {
            "name": "%s"
            }
          },
          "sort": {
          "version": {
            "order": "desc"
            }
          },
          "size": 1
        }`, name)
	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	r, e := client.Do(request)
	if e != nil {
		return
	}

	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to search latest metadata: %d", r.StatusCode)
		return
	}
	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	if len(sr.Hits.Hits) != 0 {
		meta = sr.Hits.Hits[0].Source
	}
	//log.Println(meta)
	return
}

func GetMetadata(mapping string, name string, version int) (Metadata, error) {
	if version == 0 {
		return SearchLatestVersion(mapping, name)
	}
	return getMetadata(mapping, name, version)
}

func PutMetadata(mapping string, name string, version int, size int64, hash string) error {
	time, _ := strconv.Atoi(time.Now().Format("20060102150405"))
	doc := fmt.Sprintf(`{"name":"%s","version":%d,"size":%d,"hash":"%s","time":%d}`,
		name, version, size, hash, time)

	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata_%s/_doc/%s_%d?op_type=create",
		choose(available), mapping, name, version)

	request, _ := http.NewRequest("PUT", url, strings.NewReader(doc))
	request.Header.Set("Content-Type", "application/json")
	r, e := client.Do(request)
	if e != nil {
		return e
	}
	if r.StatusCode == http.StatusConflict {
		return PutMetadata(mapping, name, version+1, size, hash)
	}
	if r.StatusCode != http.StatusCreated {
		result, _ := ioutil.ReadAll(r.Body)
		return fmt.Errorf("fail to put metadata: %d %s", r.StatusCode, string(result))
	}
	return nil
}

func AddVersion(mapping string, name, hash string, size int64) error {
	// 增加版本
	version, e := SearchLatestVersion(mapping, name)
	if e != nil {
		return e
	}
	return PutMetadata(mapping, name, version.Version+1, size, hash)
}

func SearchAllVersions(mapping string, name string, from, size int) ([]Metadata, error) {
	//url := fmt.Sprintf("http://%s/metadata/_search?sort=name,version&from=%d&size=%d",
	//	os.Getenv("ES_SERVER"), from, size)
	//if name != "" {
	//	url += "&q=name:" + name
	//}
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata_%s/_search", choose(available), mapping)
	var body string
	if name != "" {
		body = fmt.Sprintf(`
		{
			"query": {
				"match_phrase": {
					"name": "%s"
				}
			},
			"sort": {
				"version": {
					"order": "desc"
				}
			},
			"from":%d,
			"size": %d
		}
		`, name, from, size)
	} else {
		body = fmt.Sprintf(`
		{
			"sort": {
				"version": {
					"order": "desc"
				}
			},
			"from":%d,
			"size": %d
		}
		`, from, size)
	}

	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	r, e := client.Do(request)
	if e != nil {
		return nil, e
	}

	metas := make([]Metadata, 0)
	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	for i := range sr.Hits.Hits {
		metas = append(metas, sr.Hits.Hits[i].Source)
	}
	return metas, nil
}

// 自己写的去重
//func SearchObjectNum(name string) (int64, error) {
//	url := fmt.Sprintf("http://%s/metadata/_search?sort=name.subField",
//		os.Getenv("ES_SERVER"))
//	if name != "" {
//		url += "?q=name.subField:" + name // 待定
//	}
//
//	r, e := http.Get(url)
//	if e != nil {
//		return 0, e
//	}
//	metas := make([]Metadata, 0)
//	result, _ := ioutil.ReadAll(r.Body)
//	var sr searchResult
//	json.Unmarshal(result, &sr)
//	for i := range sr.Hits.Hits {
//		metas = append(metas, sr.Hits.Hits[i].Source)
//	}
//	//log.Println(metas)
//	// 去重复
//	metasResult := make([]Metadata, 0)
//	var mapVersions = make(map[string]int)
//	for i := range metas {
//		mapVersions[metas[i].Name] = metas[i].Version
//	}
//	//log.Println(mapVersions)
//	for key, value := range mapVersions {
//		for i := range metas {
//			if value == metas[i].Version && key == metas[i].Name && metas[i].Size != 0 && metas[i].Hash != "" {
//				metas[i].Name, _ = url2.QueryUnescape(metas[i].Name)
//				//log.Println(metas[i].Version)
//				metasResult = append(metasResult, metas[i])
//			}
//		}
//	}
//	//log.Println(metasResult)
//	return metasResult, nil
//}

// es的聚合去重并分页
func SearchApiVersions(mapping string, name string, from int, size int) ([]Metadata, error) {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata_%s/_search", choose(available), mapping)
	var body string
	if name != "" {
		body = fmt.Sprintf(`
		{
			"query": {
        		"match_phrase": {
					"name": "%s"
        		}
    		},
			"collapse": {
				"field": "name.subField"
			},
			"sort": [
				{
					"time":"desc"
				},
				{
					"version": "desc"
				}
			],
			"from":%d,
			"size": %d
		}
	`, name, from, size)
	} else {
		body = fmt.Sprintf(`
		{
			"query": {
				"match_all": {}
			},
			"collapse": {
				"field": "name.subField"
			},
			"sort": [
				{
					"time":"desc"
				},
				{
					"version": "desc"
				}
			],
			"from":%d,
			"size": %d
		}
	`, from, size)
	}

	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	r, e := client.Do(request)
	if e != nil {
		return nil, e
	}

	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	metas := make([]Metadata, 0)
	for _, hit := range sr.Hits.Hits {
		metas = append(metas, hit.Source)
	}
	for i := range metas {
		metas[i].Name, _ = url2.QueryUnescape(metas[i].Name)
	}
	return metas, nil
}

func DelMetadata(mapping string, name string, version int) {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata_%s/_doc/%s_%d",
		choose(available), mapping, name, version)
	request, _ := http.NewRequest("DELETE", url, nil)

	client.Do(request)
}

type Bucket struct {
	Key         string
	Doc_count   int
	Min_version struct {
		Value float32
	}
}

type aggregateResult struct {
	Aggregations struct {
		Group_by_name struct {
			Buckets []Bucket
		}
	}
}

func SearchVersionStatus(mapping string, min_doc_count int) ([]Bucket, error) {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/metadata_%s/_search", choose(available), mapping)
	body := fmt.Sprintf(`
        {
          "size": 0,
          "aggs": {
            "group_by_name": {
              "terms": {
                "field": "name.subField",
                "min_doc_count": %d
              },
              "aggs": {
                "min_version": {
                  "min": {
                    "field": "version"
                  }
                }
              }
            }
          }
        }`, min_doc_count)
	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	r, e := client.Do(request)
	if e != nil {
		return nil, e
	}
	b, _ := ioutil.ReadAll(r.Body)
	var ar aggregateResult
	json.Unmarshal(b, &ar)
	return ar.Aggregations.Group_by_name.Buckets, nil
}

func HasHash(mapping string, hash string) (bool, error) {
	url := fmt.Sprintf("http://%s/metadata_%s/_search?q=hash.subField:%s&size=0", choose(available), mapping, hash)
	r, e := http.Get(url)
	if e != nil {
		return false, e
	}
	b, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(b, &sr)
	return sr.Hits.Total.Value != 0, nil
}

func SearchHashSize(mapping string, hash string) (size int64, e error) {
	url := fmt.Sprintf("http://%s/metadata_%s/_search?q=hash:%s&size=1",
		choose(available), mapping, hash)
	r, e := http.Get(url)
	if e != nil {
		return
	}
	if r.StatusCode != http.StatusOK {
		e = fmt.Errorf("fail to search hash size: %d", r.StatusCode)
		return
	}
	result, _ := ioutil.ReadAll(r.Body)
	var sr searchResult
	json.Unmarshal(result, &sr)
	if len(sr.Hits.Hits) != 0 {
		size = sr.Hits.Hits[0].Source.Size
	}
	return
}

// PutLog 添加日志
func PutLog(doc string) {
	client := http.Client{}
	url := fmt.Sprintf("http://%s/log/_doc",
		choose(available))

	request, _ := http.NewRequest("POST", url, strings.NewReader(doc))
	request.Header.Set("Content-Type", "application/json")
	r, e := client.Do(request)
	if e != nil {
		log.Println(e)
	}
	//log.Printf("es发送code: %d", r.StatusCode)
	if r.StatusCode != http.StatusCreated {
		result, _ := ioutil.ReadAll(r.Body)
		log.Println(fmt.Errorf("fail to put log: %d %s", r.StatusCode, string(result)))
		return
	}
}

// SearchLog 获取当天某个主机 某个级别的日志
func SearchLog(searchParam map[string]interface{}, from int, size int) ([]Log, error) {
	var logData []Log             // 结果
	var requestBody searchLogBody // 请求体
	var body = ""

	if len(searchParam) > 0 { //参数为空 查询全部
		for k, v := range searchParam {
			// 如果是查询内容则使用分词
			if k == "content" {
				s := v.(string)
				requestBody.Query.Bool.Must = append(requestBody.Query.Bool.Must, map[string]map[string]string{"match": {k: s}})
			} else if k == "dateTime" { // 如果是使用时间和日期组合查询
				// s为from：时间戳 和to：时间戳
				s := v.(map[string]interface{})
				var fromDateTime float64 = 0
				var toDataTime float64 = 0
				for sk, sv := range s {
					svString := sv.(float64)
					switch sk {
					case "from":
						fromDateTime = svString
						break
					case "to":
						toDataTime = svString
						break
					}
				}
				// 组装es请求数据
				data := map[string]map[string]map[string]float64{"range": {"dateTime": {"from": fromDateTime, "to": toDataTime}}}
				requestBody.Query.Bool.Must = append(requestBody.Query.Bool.Must, data)
			} else { // 其他字段使用强制匹配
				s := v.(string)
				requestBody.Query.Bool.Must = append(requestBody.Query.Bool.Must, map[string]map[string]string{"match_phrase": {k: s}})
			}
		}
		requestBody.Sort = append(requestBody.Sort, map[string]string{"dateTime": "desc"})
		requestBody.From = from
		requestBody.Size = size
		marshal, _ := json.Marshal(requestBody)

		body = string(marshal)
	} else {
		body = fmt.Sprintf(`
		{	
			"sort": [
				{
					"dateTime": "desc"
				}
			],
			"from":%d,
			"size": %d
		}
		`, from, size)
	}

	client := http.Client{}

	url := fmt.Sprintf("http://%s/log/_search", choose(available))

	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")

	r, e := client.Do(request)
	if e != nil {
		return logData, e
	}

	if r.StatusCode != http.StatusOK {
		return logData, fmt.Errorf("查询日志失败: %d", r.StatusCode)
	}
	result, _ := ioutil.ReadAll(r.Body)

	var sr searchLogResult
	json.Unmarshal(result, &sr)

	for i := range sr.Hits.Hits {
		logData = append(logData, sr.Hits.Hits[i].Source)
	}

	return logData, nil
}
