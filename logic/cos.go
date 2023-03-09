package logic

import (
	"fmt"
	"github.com/HayWolf/osd-tool/entity"
	"github.com/HayWolf/osd-tool/helper"
	"github.com/HayWolf/osd-tool/repo/qcloud"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// QcloudCos
type QcloudCos struct {
	cosClient *cos.Client
	config    *entity.SyncConfig
}

// NewQcloudCos 实例化cosImpl
func NewQcloudCos(cfg *entity.SyncConfig) *QcloudCos {

	u, _ := url.Parse(fmt.Sprintf(
		"https://%s.cos.%s.myqcloud.com",
		cfg.Osd.Bucket,
		cfg.Osd.Region,
	),
	)

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
		config: cfg,
	}
}

// Upload 上传本地文件到Cos
func (s *QcloudCos) Upload() error {
	// 多线程执行
	keysCh := make(chan []string, 3)
	var wg sync.WaitGroup
	threads := 8
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go qcloud.Upload(&wg, s.cosClient, keysCh)
	}

	defer func() {
		// 关闭管道
		close(keysCh)
		// 等待上传完成
		wg.Wait()
	}()

	for _, dir := range s.config.Upload.List {
		log.Printf("begin to upload, from local: %s, to cos: %s", dir.Source, dir.Dest)
		err := filepath.Walk(dir.Source, func(path string, info fs.FileInfo, err error) error {
			if info == nil {
				log.Printf("no such file or directory:%s", path)
				return nil
			}

			if info.IsDir() {
				// 跳过需要忽略的文件夹
				if helper.InArray(info.Name(), s.config.Upload.Ignore) {
					log.Printf("skipping a dir:%s", path)
					return filepath.SkipDir
				}
				log.Printf("into dir:%s", path)
				return nil
			}

			// 跳过需要忽略的文件
			if helper.InArray(info.Name(), s.config.Upload.Ignore) {
				log.Printf("skipping a file:%s", path)
				return nil
			}

			// 获取 COS 中的文件路径
			cosPath := strings.Replace(path, dir.Source, dir.Dest, 1)

			// 丢进管道，异步上传
			keysCh <- []string{cosPath, path}

			return nil
		})

		if err != nil {
			log.Printf("filewalk error:%s", err.Error())
			return err
		}
	}

	return nil
}

// Download 下载Cos上的文件到本地
func (s *QcloudCos) Download() error {

	// 多线程执行
	keysCh := make(chan []string, 3)
	var wg sync.WaitGroup
	threads := 8
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go qcloud.Download(&wg, s.cosClient, keysCh)
	}

	defer func() {
		// 关闭管道
		close(keysCh)
		// 等待上传完成
		wg.Wait()
	}()

	for _, dir := range s.config.Download.List {
		log.Printf("begin to download, from cos: %s, to local: %s", dir.Source, dir.Dest)

		prefix := strings.TrimLeft(dir.Source, "/")
		objs := qcloud.ListObjects(s.cosClient, prefix, "")
		for _, c := range objs {
			source, _ := cos.DecodeURIComponent(c.Key)
			dest := strings.Replace(source, prefix, dir.Dest, 1)

			// 创建本地目录
			if _, err := os.Stat(path.Dir(dest)); err != nil && os.IsNotExist(err) {
				err := os.MkdirAll(path.Dir(dest), os.ModePerm)
				if err != nil {
					log.Printf("mkdir error:%s", err.Error())
					continue
				}
			}

			// 目录不需要下载
			if strings.HasSuffix(source, "/") {
				continue
			}
			// 丢进管道，异步下载
			keysCh <- []string{source, dest}
		}

	}
	return nil
}
