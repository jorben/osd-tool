package aliyun

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"log"
	"strings"
	"sync"
	"time"
)

var idCounter int64
var idMutex sync.Mutex

func generateID() int64 {
	idMutex.Lock()
	defer idMutex.Unlock()
	idCounter++
	return idCounter
}

// Upload 多协程上传
func Upload(wg *sync.WaitGroup, c *oss.Bucket, keysCh <-chan []string) {
	goId := generateID()
	defer wg.Done()
	for keys := range keysCh {
		key := keys[0]
		filename := keys[1]
		// 上传到OSS
		err := c.PutObjectFromFile(key, filename, nil)
		if err != nil {
			log.Printf("upload error, file:%s, error:%s", filename, err.Error())
			continue
		}
		log.Printf("upload success, gorouting:%d, file:%s", goId, key)
	}
}

// Download 多协程下载
func Download(wg *sync.WaitGroup, c *oss.Bucket, keysCh <-chan []string) {
	goId := generateID()
	defer wg.Done()
	for keys := range keysCh {
		key := keys[0]
		filename := keys[1]
		err := c.GetObjectToFile(key, filename, nil)
		if err != nil {
			log.Printf("download error, file:%s, error:%s", key, err.Error())
			continue
		}
		log.Printf("download success, gorouting:%d, file:%s", goId, filename)
	}
}

// ListObjects 获取对象列表
func ListObjects(c *oss.Bucket, prefix string, marker string) (objs []oss.ObjectProperties) {
	prefix = strings.TrimLeft(prefix, "/")
	m := oss.Marker(marker)
	i := 0
	maxRetry := 3
	isTruncated := true
	for isTruncated {
		v, err := c.ListObjects(oss.MaxKeys(10), m, oss.Prefix(prefix))
		if err != nil {
			if i < maxRetry {
				i++
				time.Sleep(time.Second)
				continue
			} else {
				log.Printf("List Objects error:%s", err.Error())
				return objs
			}
		}
		// 获取成功 重置重试次数
		i = 0
		m = oss.Marker(v.NextMarker)
		isTruncated = v.IsTruncated
		objs = append(objs, v.Objects...)
	}
	return objs
}
