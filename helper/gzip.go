package helper

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// Unarchive 解压缩tgz压缩包
func Unarchive(src string, dest string) error {
	// 打开压缩包
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	// 创建 gzip.Reader
	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	// 创建 tar.Reader
	tr := tar.NewReader(gr)

	// 解压文件
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // 结束解压
		}
		if err != nil {
			return err
		}

		// 构建目标文件路径
		target := filepath.Join(dest, hdr.Name)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.ModePerm); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.Create(target)
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
	return nil
}
