package main

import (
	"errors"
	"fmt"
	"github.com/HayWolf/osd-tool/entity"
	"github.com/HayWolf/osd-tool/helper"
	"github.com/HayWolf/osd-tool/logic"
	"github.com/ldigit/config"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"strings"
)

const VERSION = "v1.0.0"

// SyncApi 同步接口定义
type SyncApi interface {
	// Upload 执行上传
	Upload() error
	// Download 执行下载
	Download() error
}

// loadConfig 加载配置项
func loadConfig(path string) *entity.SyncConfig {
	cfg := &entity.SyncConfig{}
	if err := config.LoadAndDecode(path, cfg); err != nil {
		return nil
	}
	return cfg
}

// newSyncApi 获取SyncApi实例
func newSyncApi(cfg *entity.SyncConfig) (api SyncApi, err error) {
	storage := strings.ToLower(cfg.Storage)
	if "cos" == storage {
		api = logic.NewQcloudCos(cfg)
	} else if "oss" == storage {
		api = logic.NewAliyunOss(cfg)
	} else {
		return nil, errors.New(fmt.Sprintf("storage type '%s' is not supported", cfg.Storage))
	}
	return api, nil
}

// doMakeConfig 创建模版配置文件
func doMakeConfig(path string) error {
	// 检查文件是否存在，存在则进行备份
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		if err := os.Rename(path, path+".bak"); err != nil {
			log.Printf("Error while backup config - %s", err.Error())
			return err
		}
	}
	buf := entity.GetConfigDemo()
	dest, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dest.Close()
	if _, err := io.WriteString(dest, string(buf)); err != nil {
		return err
	}
	log.Printf("Configuration file initialization completed, look at %s\n", path)
	return nil
}

func doUpload(ctx *cli.Context) error {
	raw := config.GetGlobalConfig()
	if raw == nil {
		return errors.New("configuration is empty, please check the config file path")
	}
	cfg := raw.(*entity.SyncConfig)
	api, err := newSyncApi(cfg)
	if err != nil {
		return err
	}
	//打印上传相关配置
	fmt.Println("--------------- CONFIG ---------------")
	fmt.Println("storage:", cfg.Storage)
	fmt.Println("osd config:")
	fmt.Printf("  bucket: %s\n", cfg.Osd.Bucket)
	fmt.Println("  region:", cfg.Osd.Region)
	fmt.Println("  secret_id:", helper.HideSecret(cfg.Osd.SecretId, 8))
	fmt.Println("  secret_key:", helper.HideSecret(cfg.Osd.SecretKey, 8))
	fmt.Println("upload config:")
	fmt.Println("  ignore:", cfg.Upload.Ignore)
	fmt.Println("  list:")
	for _, path := range cfg.Upload.List {
		fmt.Printf("    %s -> %s\n", path.Source, path.Dest)
	}
	fmt.Println("--------------------------------------")
	return api.Upload()
}

func doDownload(ctx *cli.Context) error {
	raw := config.GetGlobalConfig()
	if raw == nil {
		return errors.New("configuration is empty, please check the config file path")
	}
	cfg := raw.(*entity.SyncConfig)
	api, err := newSyncApi(cfg)
	if err != nil {
		return err
	}
	// 打印下载相关配置
	fmt.Println("--------------- CONFIG ---------------")
	fmt.Println("storage:", cfg.Storage)
	fmt.Println("osd config:")
	fmt.Printf("  bucket: %s\n", cfg.Osd.Bucket)
	fmt.Println("  region:", cfg.Osd.Region)
	fmt.Println("  secret_id:", helper.HideSecret(cfg.Osd.SecretId, 8))
	fmt.Println("  secret_key:", helper.HideSecret(cfg.Osd.SecretKey, 8))
	fmt.Println("download config:")
	fmt.Println("  list:")
	for _, path := range cfg.Download.List {
		fmt.Printf("    %s -> %s\n", path.Source, path.Dest)
	}
	fmt.Println("--------------------------------------")
	return api.Download()
}

func doUpgrade(ctx *cli.Context) error {
	repoName := "HayWolf/osd-tool"
	binName := "osd-tool_{os}_{arch}"
	// 初始化实例，获取最新版本信息
	api, err := logic.NewVersionImpl(repoName, binName, binName+".tgz")
	if err != nil {
		return err
	}

	// 判断是否最新版本
	if api.IsLatest(VERSION) {
		fmt.Println("Already running the latest version.")
		return nil
	}

	// 执行升级，失败时候自动回滚
	if err := api.Update(); err != nil {
		return err
	}

	fmt.Println("Version upgrade finished, please restart.")
	return nil
}

func main() {

	// 配置文件路径，从参数获取
	var configPath string
	// 设置参数处理方法
	app := &cli.App{
		Usage:          "对象存储上传、下载工具，支持腾讯云COS、阿里云OSS",
		DefaultCommand: "",
		Version:        VERSION,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				DefaultText: "config.yaml",
				FilePath:    "",
				Usage:       "配置文件路径",
				Destination: &configPath,
				Value:       "config.yaml",
			},
			&cli.BoolFlag{
				Name:               "upgrade",
				Usage:              "检查和升级当前工具版本",
				DisableDefaultText: true,
				Action: func(cCtx *cli.Context, b bool) error {
					err := doUpgrade(cCtx)
					if err == nil {
						os.Exit(0)
					}
					return err
				},
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "upload",
				Aliases: []string{"u"},
				Usage:   "把配置的本地目录上传到云端对象存储中",
				Action:  doUpload,
			},
			{
				Name:    "download",
				Aliases: []string{"d"},
				Usage:   "按配置从云端对象存储中下载文件到本地",
				Action:  doDownload,
			},
			{
				Name:    "init",
				Aliases: []string{"i"},
				Usage:   "按模版初始化配置文件",
				Action: func(cCtx *cli.Context) error {
					return doMakeConfig(configPath)
				},
			},
		},
		Before: func(cCtx *cli.Context) error {
			// 初始化配置内容，存储到全局变量中
			if cfg := loadConfig(configPath); cfg != nil {
				config.SetGlobalConfig(cfg)
			}
			return nil
		},
		CommandNotFound: func(cCtx *cli.Context, command string) {
			fmt.Fprintf(cCtx.App.Writer, "不支持这个指令%q，请通过help指令查看使用方法\n", command)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

	return
}
