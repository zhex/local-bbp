package common

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

func IsFileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func Untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	tarReader := tar.NewReader(file)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := dest + "/" + header.Name
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			fileToWrite, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(fileToWrite, tarReader); err != nil {
				_ = fileToWrite.Close()
				return err
			}
			_ = fileToWrite.Close()
		default:
			continue
		}
	}

	return nil
}
