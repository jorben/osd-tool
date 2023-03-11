package provider

import (
	"context"
	"fmt"
	"github.com/HayWolf/osd-tool/config"
	"github.com/tencentyun/cos-go-sdk-v5"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// QcloudCos
type QcloudCos struct {
	cosClient *cos.Client
}

// NewQcloudCos 实例化cosImpl
func NewQcloudCos(cfg *config.TransferConfig) *QcloudCos {

	u, _ := url.Parse(fmt.Sprintf(
		"https://%s.cos.%s.myqcloud.com",
		cfg.Osd.Bucket,
		cfg.Osd.Region,
	))

	return &QcloudCos{
		cosClient: cos.NewClient(
			&cos.BaseURL{BucketURL: u},
			&http.Client{
				Timeout: time.Second * time.Duration(cfg.Osd.Timeout),
				Transport: &cos.AuthorizationTransport{
					SecretID:  cfg.Osd.SecretId,
					SecretKey: cfg.Osd.SecretKey,
				},
			},
		),
	}
}

func (s *QcloudCos) GetFile(key string, filepath string) error {
	_, err := s.cosClient.Object.GetToFile(context.Background(), key, filepath, nil)
	if err != nil {
		log.Printf("GetToFile error, file:%s, error:%s", key, err.Error())
	}
	return err
}

func (s *QcloudCos) PutFile(key string, filepath string) error {
	_, err := s.cosClient.Object.PutFromFile(context.Background(), key, filepath, nil)
	if err != nil {
		log.Printf("PutFromFile error, file:%s, error:%s", filepath, err.Error())
	}
	return err
}

func (s *QcloudCos) List(prefix string, marker string) (list []string) {
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
		v, _, err := s.cosClient.Bucket.Get(context.Background(), opt)
		if err != nil {
			if i < maxRetry {
				i++
				time.Sleep(time.Second)
				continue
			} else {
				log.Printf("Get Bucket error:%s", err.Error())
				return list
			}
		}

		for _, c := range v.Contents {
			source, _ := cos.DecodeURIComponent(c.Key)
			list = append(list, source)
		}

		// 获取成功重置重试次数
		i = 0
		marker = v.Marker
		isTruncated = v.IsTruncated
	}
	return list
}
