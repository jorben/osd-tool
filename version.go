package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jorben/osd-tool/helper"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Updater 版本更新器
type Updater struct {
	latest   *Release
	repoName string
	pkgName  string
	binName  string
}

// Release API response 结构
type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		Url  string `json:"browser_download_url"`
	} `json:"assets"`
}

// NewUpdater 获取Version实例
func NewUpdater(repo string, bin string, pkg string) (*Updater, error) {
	latest, err := getLatest(repo)
	if err != nil {
		return nil, err
	}

	return &Updater{
		latest:   &latest,
		repoName: repo,
		pkgName:  strings.Replace(strings.Replace(pkg, "{arch}", runtime.GOARCH, -1), "{os}", runtime.GOOS, -1),
		binName:  strings.Replace(strings.Replace(bin, "{arch}", runtime.GOARCH, -1), "{os}", runtime.GOOS, -1),
	}, nil
}

// Upgrade 执行版本升级
func (s *Updater) Upgrade() error {
	url := s.getLatestUrl()
	if url == "" {
		return errors.New(
			fmt.Sprintf("The latest version does not support your system: %s-%s", runtime.GOOS, runtime.GOARCH))
	}

	// 创建本地临时目录
	tmpPath := fmt.Sprintf("%s/.%s/", os.TempDir(), s.repoName)
	if _, err := os.Stat(tmpPath); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(tmpPath, os.ModePerm); err != nil {
			return err
		}
	}
	defer os.RemoveAll(tmpPath)

	// 下载包到本地
	pkgPath := filepath.Join(tmpPath, s.pkgName)
	binPath := filepath.Join(tmpPath, s.binName)

	fd, err := os.OpenFile(pkgPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		return err
	}
	defer fd.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bar := progressbar.DefaultBytes(resp.ContentLength, "downloading...")

	if _, err := io.Copy(io.MultiWriter(fd, bar), resp.Body); err != nil {
		return err
	}

	// 解压压缩包
	if err := helper.Unarchive(pkgPath, tmpPath); err != nil {
		return err
	}

	// 判断文件是否存在
	if _, err := os.Stat(binPath); err != nil && os.IsNotExist(err) {
		return err
	}

	// 替换文件
	self, err := os.Executable()
	if err != nil {
		return err
	}

	// 必须先rename 否则替换后无法执行
	if err := os.Rename(self, self+".bak"); err != nil {
		return err
	}
	defer os.Remove(self + ".bak")

	if err := helper.Copy(binPath, self); err != nil {
		rollback(self+".bak", self)
		return err
	}

	// 修改目标文件的权限
	if err := os.Chmod(self, 0755); err != nil {
		rollback(self+".bak", self)
		return err
	}
	return nil
}

// IsLatest 判断当前版本是否最新版
func (s *Updater) IsLatest(currVersion string) bool {
	if 1 == helper.CompareVersion(s.latest.TagName, currVersion) {
		return false
	}
	return true
}

// GetLatestUrl 获取与系统匹配的版本链接
func (s *Updater) getLatestUrl() string {
	for _, pkg := range s.latest.Assets {
		if s.pkgName == pkg.Name {
			return pkg.Url
		}
	}
	return ""
}

func rollback(src string, dest string) {
	log.Println("Upgrade failed, rolled back")
	// 回滚
	if err := os.Rename(src, dest); err != nil {
		log.Printf("Error while rollback - %s", err.Error())
		return
	}
	return
}

// getApi 获取API地址
func getApi(repo string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
}

// getLatest 获取最新版本信息
func getLatest(repo string) (Release, error) {
	latest := Release{}
	res, err := http.Get(getApi(repo))
	if err != nil {
		log.Println("Error getting latest release from Github:", err)
		return latest, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println("Error reading latest release from Github:", err)
		return latest, err
	}

	if err := json.Unmarshal(body, &latest); err != nil {
		log.Println("Error decoding latest release from Github:", err)
		return latest, err
	}

	return latest, nil
}
