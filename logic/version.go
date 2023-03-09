package logic

import (
	"errors"
	"fmt"
	"github.com/HayWolf/osd-tool/helper"
	"github.com/HayWolf/osd-tool/repo/github"
	"github.com/schollz/progressbar/v3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type VersionImpl struct {
	latest   *github.Release
	repoName string
	pkgName  string
	binName  string
}

// NewVersionImpl 获取Version实例
func NewVersionImpl(repo string, bin string, pkg string) (*VersionImpl, error) {
	latest, err := github.GetLatest(repo)
	if err != nil {
		return nil, err
	}

	return &VersionImpl{
		latest:   &latest,
		repoName: repo,
		pkgName:  strings.Replace(strings.Replace(pkg, "{arch}", runtime.GOARCH, -1), "{os}", runtime.GOOS, -1),
		binName:  strings.Replace(strings.Replace(bin, "{arch}", runtime.GOARCH, -1), "{os}", runtime.GOOS, -1),
	}, nil
}

// Update 执行版本升级
func (s *VersionImpl) Update() error {
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
func (s *VersionImpl) IsLatest(currVersion string) bool {
	if 1 == helper.CompareVersion(s.latest.TagName, currVersion) {
		return false
	}
	return true
}

// GetLatestUrl 获取与系统匹配的版本链接
func (s *VersionImpl) getLatestUrl() string {
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
