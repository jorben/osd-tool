package qcloud

import (
	"context"
	"github.com/tencentyun/cos-go-sdk-v5"
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

// Upload 多线程上传
func Upload(wg *sync.WaitGroup, c *cos.Client, keysCh <-chan []string) {
	goId := generateID()
	defer wg.Done()
	for keys := range keysCh {
		key := keys[0]
		filename := keys[1]
		// 上传到COS
		_, err := c.Object.PutFromFile(context.Background(), key, filename, nil)
		if err != nil {
			log.Printf("upload error, file:%s, error:%s", filename, err.Error())
			continue
		}
		log.Printf("upload success, gorouting:%d, file:%s", goId, key)
	}
}

// Download 多线程下载
func Download(wg *sync.WaitGroup, c *cos.Client, keysCh <-chan []string) {
	goId := generateID()
	defer wg.Done()
	for keys := range keysCh {
		key := keys[0]
		filename := keys[1]
		_, err := c.Object.GetToFile(context.Background(), key, filename, nil)
		if err != nil {
			log.Printf("download error, file:%s, error:%s", key, err.Error())
			continue
		}
		log.Printf("download success, gorouting:%d, file:%s", goId, filename)
	}
}

// ListObjects 获取对象列表
func ListObjects(c *cos.Client, prefix string, marker string) (objs []cos.Object) {
	prefix = strings.TrimLeft(prefix, "/")
	i := 0
	maxRetry := 3
	isTruncated := true
	for isTruncated {
		opt := &cos.BucketGetOptions{
			Prefix:       prefix,
			Marker:       marker,
			EncodingType: "url", // url编码
		}
		v, _, err := c.Bucket.Get(context.Background(), opt)
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

		// 获取成功重置重试次数
		i = 0
		marker = v.Marker
		isTruncated = v.IsTruncated
		objs = append(objs, v.Contents...)
	}
	return objs
}
