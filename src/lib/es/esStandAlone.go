package es

//
//import (
//	"encoding/json"
//	"fmt"
//	"io/ioutil"
//	"net/http"
//	url2 "net/url"
//	"os"
//	"strconv"
//	"strings"
//	"time"
//)
//
//type Metadata struct {
//	Name    string
//	Version int
//	Size    int64
//	Hash    string
//	Time    int64
//}
//
//type hit struct {
//	Source Metadata `json:"_source"`
//}
//
//type searchResult struct {
//	Hits struct {
//		Total struct {
//			Value    int
//			Relation string
//		}
//		Hits []hit
//	}
//}
//
//func getMetadata(name string, versionId int) (meta Metadata, e error) {
//	url := fmt.Sprintf("http://%s/metadata/_doc/%s_%d/_source",
//		os.Getenv("ES_SERVER"), name, versionId)
//	r, e := http.Get(url)
//	if e != nil {
//		return
//	}
//	if r.StatusCode != http.StatusOK {
//		e = fmt.Errorf("fail to get %s_%d: %d", name, versionId, r.StatusCode)
//		return
//	}
//	result, _ := ioutil.ReadAll(r.Body)
//	json.Unmarshal(result, &meta)
//	return
//}
//
//func SearchLatestVersion(name string) (meta Metadata, e error) {
//	client := http.Client{}
//	//url := fmt.Sprintf("http://%s/metadata/_search?q=name:%s&size=1&sort=version:desc",
//	//	os.Getenv("ES_SERVER"),name)
//	url := fmt.Sprintf("http://%s/metadata/_search", os.Getenv("ES_SERVER"))
//	body := fmt.Sprintf(`
//		{
//  		  "query": {
//            "match_phrase": {
//            "name": "%s"
//            }
//          },
//          "sort": {
//          "version": {
//            "order": "desc"
//            }
//          },
//          "size": 1
//        }`, name)
//	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
//	request.Header.Set("Content-Type", "application/json")
//
//	r, e := client.Do(request)
//	if e != nil {
//		return
//	}
//
//	if r.StatusCode != http.StatusOK {
//		e = fmt.Errorf("fail to search latest metadata: %d", r.StatusCode)
//		return
//	}
//	result, _ := ioutil.ReadAll(r.Body)
//	var sr searchResult
//	json.Unmarshal(result, &sr)
//	if len(sr.Hits.Hits) != 0 {
//		meta = sr.Hits.Hits[0].Source
//	}
//	//log.Println(meta)
//	return
//}
//
//func GetMetadata(name string, version int) (Metadata, error) {
//	if version == 0 {
//		return SearchLatestVersion(name)
//	}
//	return getMetadata(name, version)
//}
//
//func PutMetadata(name string, version int, size int64, hash string) error {
//	time, _ := strconv.Atoi(time.Now().Format("20060102150405"))
//	doc := fmt.Sprintf(`{"name":"%s","version":%d,"size":%d,"hash":"%s","time":%d}`,
//		name, version, size, hash, time)
//
//	client := http.Client{}
//	url := fmt.Sprintf("http://%s/metadata/_doc/%s_%d?op_type=create",
//		os.Getenv("ES_SERVER"), name, version)
//
//	request, _ := http.NewRequest("PUT", url, strings.NewReader(doc))
//	request.Header.Set("Content-Type", "application/json")
//	r, e := client.Do(request)
//	if e != nil {
//		return e
//	}
//	if r.StatusCode == http.StatusConflict {
//		return PutMetadata(name, version+1, size, hash)
//	}
//	if r.StatusCode != http.StatusCreated {
//		result, _ := ioutil.ReadAll(r.Body)
//		return fmt.Errorf("fail to put metadata: %d %s", r.StatusCode, string(result))
//	}
//	return nil
//}
//
//func AddVersion(name, hash string, size int64) error {
//	// 增加版本
//	version, e := SearchLatestVersion(name)
//	if e != nil {
//		return e
//	}
//	return PutMetadata(name, version.Version+1, size, hash)
//}
//
//func SearchAllVersions(name string, from, size int) ([]Metadata, error) {
//	//url := fmt.Sprintf("http://%s/metadata/_search?sort=name,version&from=%d&size=%d",
//	//	os.Getenv("ES_SERVER"), from, size)
//	//if name != "" {
//	//	url += "&q=name:" + name
//	//}
//	client := http.Client{}
//	url := fmt.Sprintf("http://%s/metadata/_search", os.Getenv("ES_SERVER"))
//	var body string
//	if name != "" {
//		body = fmt.Sprintf(`
//		{
//			"query": {
//				"match_phrase": {
//					"name": "%s"
//				}
//			},
//			"sort": {
//				"version": {
//					"order": "desc"
//				}
//			},
//			"from":%d,
//			"size": %d
//		}
//		`, name, from, size)
//	} else {
//		body = fmt.Sprintf(`
//		{
//			"sort": {
//				"version": {
//					"order": "desc"
//				}
//			},
//			"from":%d,
//			"size": %d
//		}
//		`, from, size)
//	}
//
//	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
//	request.Header.Set("Content-Type", "application/json")
//
//	r, e := client.Do(request)
//	if e != nil {
//		return nil, e
//	}
//
//	metas := make([]Metadata, 0)
//	result, _ := ioutil.ReadAll(r.Body)
//	var sr searchResult
//	json.Unmarshal(result, &sr)
//	for i := range sr.Hits.Hits {
//		metas = append(metas, sr.Hits.Hits[i].Source)
//	}
//	return metas, nil
//}
//
//// 自己写的去重
////func SearchObjectNum(name string) (int64, error) {
////	url := fmt.Sprintf("http://%s/metadata/_search?sort=name.subField",
////		os.Getenv("ES_SERVER"))
////	if name != "" {
////		url += "?q=name.subField:" + name // 待定
////	}
////
////	r, e := http.Get(url)
////	if e != nil {
////		return 0, e
////	}
////	metas := make([]Metadata, 0)
////	result, _ := ioutil.ReadAll(r.Body)
////	var sr searchResult
////	json.Unmarshal(result, &sr)
////	for i := range sr.Hits.Hits {
////		metas = append(metas, sr.Hits.Hits[i].Source)
////	}
////	//log.Println(metas)
////	// 去重复
////	metasResult := make([]Metadata, 0)
////	var mapVersions = make(map[string]int)
////	for i := range metas {
////		mapVersions[metas[i].Name] = metas[i].Version
////	}
////	//log.Println(mapVersions)
////	for key, value := range mapVersions {
////		for i := range metas {
////			if value == metas[i].Version && key == metas[i].Name && metas[i].Size != 0 && metas[i].Hash != "" {
////				metas[i].Name, _ = url2.QueryUnescape(metas[i].Name)
////				//log.Println(metas[i].Version)
////				metasResult = append(metasResult, metas[i])
////			}
////		}
////	}
////	//log.Println(metasResult)
////	return metasResult, nil
////}
//
//// es的聚合去重并分页
//func SearchApiVersions(name string) ([]Metadata, error) {
//	client := http.Client{}
//	url := fmt.Sprintf("http://%s/metadata/_search", os.Getenv("ES_SERVER"))
//	var body string
//	if name != "" {
//		body = fmt.Sprintf(`
//		{
//			"query": {
//				"match": {
//					"name": "%s"
//				}
//			},
//			"collapse": {
//				"field": "name.subField"
//			},
//			"sort": [
//				{
//					"time": "desc"
//				},
//				{
//					"version": "desc"
//				}
//			]
//		}
//	`, name)
//	} else {
//		body = fmt.Sprintf(`
//		{
//			"query": {
//				"match_all": {}
//			},
//			"collapse": {
//				"field": "name.subField"
//			},
//			"sort": [
//				{
//					"time":"desc"
//				},
//				{
//					"version": "desc"
//				}
//			]
//		}
//	`)
//	}
//
//	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
//	request.Header.Set("Content-Type", "application/json")
//
//	r, e := client.Do(request)
//	if e != nil {
//		return nil, e
//	}
//
//	result, _ := ioutil.ReadAll(r.Body)
//
//	var sr searchResult
//	json.Unmarshal(result, &sr)
//
//	metas := make([]Metadata, 0)
//	for _, hit := range sr.Hits.Hits {
//		metas = append(metas, hit.Source)
//	}
//	for i := range metas {
//		metas[i].Name, _ = url2.QueryUnescape(metas[i].Name)
//	}
//	return metas, nil
//}
//
//func DelMetadata(name string, version int) {
//	client := http.Client{}
//	url := fmt.Sprintf("http://%s/metadata/_doc/%s_%d",
//		os.Getenv("ES_SERVER"), name, version)
//	request, _ := http.NewRequest("DELETE", url, nil)
//
//	client.Do(request)
//}
//
//type Bucket struct {
//	Key         string
//	Doc_count   int
//	Min_version struct {
//		Value float32
//	}
//}
//
//type aggregateResult struct {
//	Aggregations struct {
//		Group_by_name struct {
//			Buckets []Bucket
//		}
//	}
//}
//
//func SearchVersionStatus(min_doc_count int) ([]Bucket, error) {
//	client := http.Client{}
//	url := fmt.Sprintf("http://%s/metadata/_search", os.Getenv("ES_SERVER"))
//	body := fmt.Sprintf(`
//        {
//          "size": 0,
//          "aggs": {
//            "group_by_name": {
//              "terms": {
//                "field": "name.subField",
//                "min_doc_count": %d
//              },
//              "aggs": {
//                "min_version": {
//                  "min": {
//                    "field": "version"
//                  }
//                }
//              }
//            }
//          }
//        }`, min_doc_count)
//	request, _ := http.NewRequest("GET", url, strings.NewReader(body))
//	request.Header.Set("Content-Type", "application/json")
//	r, e := client.Do(request)
//	if e != nil {
//		return nil, e
//	}
//	b, _ := ioutil.ReadAll(r.Body)
//	var ar aggregateResult
//	json.Unmarshal(b, &ar)
//	return ar.Aggregations.Group_by_name.Buckets, nil
//}
//
//func HasHash(hash string) (bool, error) {
//	url := fmt.Sprintf("http://%s/metadata/_search?q=hash.subField:%s&size=0", os.Getenv("ES_SERVER"), hash)
//	r, e := http.Get(url)
//	if e != nil {
//		return false, e
//	}
//	b, _ := ioutil.ReadAll(r.Body)
//	var sr searchResult
//	json.Unmarshal(b, &sr)
//	return sr.Hits.Total.Value != 0, nil
//}
//
//func SearchHashSize(hash string) (size int64, e error) {
//	url := fmt.Sprintf("http://%s/metadata/_search?q=hash:%s&size=1",
//		os.Getenv("ES_SERVER"), hash)
//	r, e := http.Get(url)
//	if e != nil {
//		return
//	}
//	if r.StatusCode != http.StatusOK {
//		e = fmt.Errorf("fail to search hash size: %d", r.StatusCode)
//		return
//	}
//	result, _ := ioutil.ReadAll(r.Body)
//	var sr searchResult
//	json.Unmarshal(result, &sr)
//	if len(sr.Hits.Hits) != 0 {
//		size = sr.Hits.Hits[0].Source.Size
//	}
//	return
//}
