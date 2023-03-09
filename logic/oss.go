package logic

import (
	"fmt"
	"github.com/HayWolf/osd-tool/entity"
	"github.com/HayWolf/osd-tool/helper"
	"github.com/HayWolf/osd-tool/repo/aliyun"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

type AliyunOss struct {
	ossBucket *oss.Bucket
	config    *entity.SyncConfig
}

// NewAliyunOss 实例化ossImpl
func NewAliyunOss(cfg *entity.SyncConfig) *AliyunOss {

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
		config:    cfg,
	}
}

// Upload 上传本地文件到Cos
func (s *AliyunOss) Upload() error {
	// 多线程执行
	keysCh := make(chan []string, 3)
	var wg sync.WaitGroup
	threads := 8
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go aliyun.Upload(&wg, s.ossBucket, keysCh)
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
			cosPath := strings.TrimLeft(strings.Replace(path, dir.Source, dir.Dest, 1), "/")

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
func (s *AliyunOss) Download() error {

	// 多线程执行
	keysCh := make(chan []string, 3)
	var wg sync.WaitGroup
	threads := 8
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go aliyun.Download(&wg, s.ossBucket, keysCh)
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
		objs := aliyun.ListObjects(s.ossBucket, prefix, "")

		for _, c := range objs {
			dest := strings.Replace(c.Key, prefix, dir.Dest, 1)
			// 创建本地目录
			if _, err := os.Stat(path.Dir(dest)); err != nil && os.IsNotExist(err) {
				err := os.MkdirAll(path.Dir(dest), os.ModePerm)
				if err != nil {
					log.Printf("mkdir error:%s", err.Error())
					continue
				}
			}

			// 目录不需要下载
			if strings.HasSuffix(c.Key, "/") {
				continue
			}

			// 丢进管道，异步下载
			keysCh <- []string{c.Key, dest}
		}

	}
	return nil
}
