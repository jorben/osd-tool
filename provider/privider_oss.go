package provider

import (
	"fmt"
	"github.com/HayWolf/osd-tool/config"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"log"
	"strings"
	"time"
)

type AliyunOss struct {
	ossBucket *oss.Bucket
}

// NewAliyunOss 实例化ossImpl
func NewAliyunOss(cfg *config.TransferConfig) *AliyunOss {

	client, err := oss.New(
		fmt.Sprintf("https://oss-%s.aliyuncs.com", cfg.Osd.Region), cfg.Osd.SecretId, cfg.Osd.SecretKey)
	if err != nil {
		log.Fatalln("new oss error:", err.Error())
	}

	bucket, err := client.Bucket(cfg.Osd.Bucket)
	if err != nil {
		log.Fatalln("new bucket error:", err.Error())
	}

	return &AliyunOss{
		ossBucket: bucket,
	}
}

func (s *AliyunOss) GetFile(key string, filepath string) error {

	err := s.ossBucket.GetObjectToFile(key, filepath, nil)
	if err != nil {
		log.Printf("GetObjectToFile error, file:%s, error:%s", key, err.Error())
	}
	return err
}

func (s *AliyunOss) PutFile(key string, filepath string) error {
	err := s.ossBucket.PutObjectFromFile(key, filepath, nil)
	if err != nil {
		log.Printf("PutObjectFromFile error, file:%s, error:%s", filepath, err.Error())
	}
	return err
}

func (s *AliyunOss) List(prefix string, marker string) (list []string) {
	prefix = strings.TrimLeft(prefix, "/")
	m := oss.Marker(marker)
	i := 0
	maxRetry := 3
	isTruncated := true
	for isTruncated {
		v, err := s.ossBucket.ListObjects(oss.MaxKeys(10), m, oss.Prefix(prefix))
		if err != nil {
			if i < maxRetry {
				i++
				time.Sleep(time.Second)
				continue
			} else {
				log.Printf("ListObjects error:%s", err.Error())
				return list
			}
		}
		for _, c := range v.Objects {
			list = append(list, c.Key)
		}
		// 获取成功 重置重试次数
		i = 0
		m = oss.Marker(v.NextMarker)
		isTruncated = v.IsTruncated
	}
	return list
}
