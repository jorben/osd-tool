package helper

import (
	"io"
	"os"
)

// Copy copies from src to dest
func Copy(src string, dest string) error {
	srcFd, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFd.Close()
	dstFd, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dstFd.Close()
	if _, err := io.Copy(dstFd, srcFd); err != nil {
		return err
	}
	return nil
}
