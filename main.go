package main

import (
	"errors"
	"fmt"
	"github.com/HayWolf/osd-tool/config"
	conf "github.com/ldigit/config"
	"github.com/urfave/cli/v2"
	"io"
	"log"
	"os"
	"strings"
)

const Version = "v1.0.2"
const RepoName = "HayWolf/osd-tool"
const BinName = "osd-tool_{os}_{arch}"
const PackageName = "osd-tool_{os}_{arch}.tgz"

// loadConfig 加载配置项
func loadConfig(path string) *config.TransferConfig {
	cfg := &config.TransferConfig{}
	if err := conf.LoadAndDecode(path, cfg); err != nil {
		return nil
	}
	// 处理Path中带有~的情况，替换为真实绝对路径
	homeDir, _ := os.UserHomeDir()
	for i, p := range cfg.Upload.List {
		if len(p.Source) > 0 && "~" == p.Source[0:1] {
			cfg.Upload.List[i].Source = strings.Replace(p.Source, "~", homeDir, 1)
		}
	}
	for i, p := range cfg.Download.List {
		if len(p.Dest) > 0 && "~" == p.Dest[0:1] {
			cfg.Download.List[i].Dest = strings.Replace(p.Dest, "~", homeDir, 1)
		}
	}
	return cfg
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
	buf := config.GetConfigDemo()
	dest, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dest.Close()
	if _, err := io.WriteString(dest, string(buf)); err != nil {
		return err
	}
	fmt.Printf("Configuration file initialization completed, look at %s\n", path)
	return nil
}

// doUpload 执行上传
func doUpload(ctx *cli.Context) error {
	raw := conf.GetGlobalConfig()
	if raw == nil {
		return errors.New("configuration is empty, please check the config file path")
	}
	cfg := raw.(*config.TransferConfig)
	transfer, err := NewTransfer(cfg)
	if err != nil {
		return err
	}
	return transfer.Upload()
}

// doDownload 执行下载
func doDownload(ctx *cli.Context) error {
	raw := conf.GetGlobalConfig()
	if raw == nil {
		return errors.New("configuration is empty, please check the config file path")
	}
	cfg := raw.(*config.TransferConfig)
	transfer, err := NewTransfer(cfg)
	if err != nil {
		return err
	}
	return transfer.Download()
}

// doUpgrade 执行当前程序的版本升级
func doUpgrade(ctx *cli.Context) error {
	// 初始化实例，获取最新版本信息
	updater, err := NewUpdater(RepoName, BinName, PackageName)
	if err != nil {
		return err
	}

	// 判断是否最新版本
	if updater.IsLatest(Version) {
		fmt.Println("Already running the latest version.")
		return nil
	}

	// 执行升级，失败时候自动回滚
	if err := updater.Upgrade(); err != nil {
		return err
	}

	fmt.Println("Version upgrade finished, please restart.")
	return nil
}

func main() {

	// 配置文件路径，从参数获取
	var configPath string
	// 支持的指令
	commends := []*cli.Command{
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
	}

	// 支持的参数
	flags := []cli.Flag{
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
	}
	// 设置参数处理方法
	app := &cli.App{
		Usage:          "对象存储上传、下载工具，支持腾讯云COS、阿里云OSS",
		DefaultCommand: "",
		Version:        Version,
		Flags:          flags,
		Commands:       commends,
		Before: func(cCtx *cli.Context) error {
			// 初始化配置内容，存储到全局变量中
			if cfg := loadConfig(configPath); cfg != nil {
				conf.SetGlobalConfig(cfg)
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
