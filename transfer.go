package main

import (
	"errors"
	"fmt"
	"github.com/jorben/osd-tool/config"
	"github.com/jorben/osd-tool/helper"
	"github.com/jorben/osd-tool/provider"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// CloudTransfer 对象存储文件传输器
type CloudTransfer struct {
	Provider provider.Provider
	Config   *config.TransferConfig
}

// NewTransfer 获取CloudTransfer实例
func NewTransfer(cfg *config.TransferConfig) (transfer *CloudTransfer, err error) {
	transfer = &CloudTransfer{Config: cfg}
	storage := strings.ToLower(cfg.Storage)
	switch storage {
	case provider.COS:
		transfer.Provider = provider.NewQcloudCos(cfg)
	case provider.OSS:
		transfer.Provider = provider.NewAliyunOss(cfg)
	default:
		return nil, errors.New(fmt.Sprintf("storage type '%s' is not supported", cfg.Storage))
	}
	return transfer, nil
}

// Upload 上传本地配置的文件目录到云端对象存储
func (t *CloudTransfer) Upload() error {
	t.PrintUploadConfig()
	// 多线程执行
	keysCh := make(chan []string, 8)
	var wg sync.WaitGroup
	threads := 8
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go t.AsyncUpload(&wg, keysCh)
	}

	defer func() {
		// 关闭管道
		close(keysCh)
		// 等待上传完成
		wg.Wait()
	}()

	for _, dir := range t.Config.Upload.List {
		log.Printf("begin to upload, from local: %s, to osd: %s", dir.Source, dir.Dest)
		err := filepath.Walk(dir.Source, func(path string, info fs.FileInfo, err error) error {
			if info == nil {
				log.Printf("no such file or directory: %s", path)
				return nil
			}

			if info.IsDir() {
				// 跳过需要忽略的文件夹
				if helper.InArray(info.Name(), t.Config.Upload.Ignore) {
					log.Printf("skipping a dir: %s", path)
					return filepath.SkipDir
				}
				log.Printf("into dir:%s", path)
				return nil
			}

			// 跳过需要忽略的文件
			if helper.InArray(info.Name(), t.Config.Upload.Ignore) {
				log.Printf("skipping a file:%s", path)
				return nil
			}

			// 获取 osd 中的文件路径
			//osdPath := strings.Replace(path, dir.Source, dir.Dest, 1)
			osdPath := strings.TrimLeft(strings.Replace(path, dir.Source, dir.Dest, 1), "/")

			// 丢进管道，异步上传
			keysCh <- []string{osdPath, path}

			return nil
		})

		if err != nil {
			log.Printf("filewalk error:%s", err.Error())
			return err
		}
	}
	return nil
}

// Download 下载配置的云端对象存储的文件到本地
func (t *CloudTransfer) Download() error {
	t.PrintDownloadConfig()
	// 多线程执行
	keysCh := make(chan []string, 8)
	var wg sync.WaitGroup
	threads := 8
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go t.AsyncDownload(&wg, keysCh)
	}

	defer func() {
		// 关闭管道
		close(keysCh)
		// 等待上传完成
		wg.Wait()
	}()

	for _, dir := range t.Config.Download.List {
		log.Printf("begin to download, from osd: %s, to local: %s", dir.Source, dir.Dest)

		prefix := strings.TrimLeft(dir.Source, "/")
		objs := t.Provider.List(prefix, "")
		for _, source := range objs {
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

// AsyncUpload 多协程上传
func (t *CloudTransfer) AsyncUpload(wg *sync.WaitGroup, keysCh <-chan []string) {
	defer wg.Done()
	for keys := range keysCh {
		key := keys[0]
		filename := keys[1]
		// 上传到对象存储
		err := t.Provider.PutFile(key, filename)
		if err != nil {
			continue
		}
		log.Printf("upload success, file:%s", key)
	}
}

// AsyncDownload 多协程下载
func (t *CloudTransfer) AsyncDownload(wg *sync.WaitGroup, ch <-chan []string) {
	defer wg.Done()
	for keys := range ch {
		key := keys[0]
		filename := keys[1]
		err := t.Provider.GetFile(key, filename)
		if err != nil {
			continue
		}
		log.Printf("download success, file:%s", filename)
	}
}

// PrintUploadConfig 打印上传相关配置
func (t *CloudTransfer) PrintUploadConfig() {
	fmt.Println("--------------- CONFIG ---------------")
	fmt.Println("storage:", t.Config.Storage)
	fmt.Println("osd config:")
	fmt.Printf("  bucket: %s\n", t.Config.Osd.Bucket)
	fmt.Println("  region:", t.Config.Osd.Region)
	fmt.Println("  secret_id:", helper.HideSecret(t.Config.Osd.SecretId, 8))
	fmt.Println("  secret_key:", helper.HideSecret(t.Config.Osd.SecretKey, 8))
	fmt.Println("upload config:")
	fmt.Println("  ignore:", t.Config.Upload.Ignore)
	fmt.Println("  list:")
	for _, p := range t.Config.Upload.List {
		fmt.Printf("    %s -> %s\n", p.Source, p.Dest)
	}
	fmt.Println("--------------------------------------")
}

// PrintDownloadConfig 打印下载相关配置
func (t *CloudTransfer) PrintDownloadConfig() {
	fmt.Println("--------------- CONFIG ---------------")
	fmt.Println("storage:", t.Config.Storage)
	fmt.Println("osd config:")
	fmt.Printf("  bucket: %s\n", t.Config.Osd.Bucket)
	fmt.Println("  region:", t.Config.Osd.Region)
	fmt.Println("  secret_id:", helper.HideSecret(t.Config.Osd.SecretId, 8))
	fmt.Println("  secret_key:", helper.HideSecret(t.Config.Osd.SecretKey, 8))
	fmt.Println("download config:")
	fmt.Println("  list:")
	for _, p := range t.Config.Download.List {
		fmt.Printf("    %s -> %s\n", p.Source, p.Dest)
	}
	fmt.Println("--------------------------------------")
}
