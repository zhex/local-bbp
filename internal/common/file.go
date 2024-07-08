package common

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func IsFileExists(path string) bool {
	f, err := os.Stat(path)
	return !os.IsNotExist(err) && !f.IsDir()
}

func IsDirExists(path string) bool {
	f, err := os.Stat(path)
	return !os.IsNotExist(err) && f.IsDir()
}

func ExtractTarFromFile(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	return ExtractTar(file, dest)
}

func ExtractTarGz(gzipStream io.Reader, dest string) error {
	stream, err := gzip.NewReader(gzipStream)
	if err != nil {
		return err
	}
	defer stream.Close()

	return ExtractTar(stream, dest)
}

func ExtractTar(reader io.Reader, dest string) error {
	tarReader := tar.NewReader(reader)
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
			fileToWrite, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
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

func GetFileSha256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func GetFilesSha256(paths []string) (string, error) {
	count := len(paths)

	if count == 0 {
		return "", nil
	}

	if count == 1 {
		return GetFileSha256(paths[0])
	}

	type result struct {
		index int
		sha   string
		err   error
	}

	results := make(chan result, count)
	var wg sync.WaitGroup

	for i, path := range paths {
		wg.Add(1)
		go func(i int, path string) {
			defer wg.Done()
			sha, err := GetFileSha256(path)
			results <- result{i, sha, err}
		}(i, path)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var keys = make([]string, count)
	for res := range results {
		if res.err != nil {
			return "", res.err
		}
		keys[res.index] = res.sha
	}

	h := sha256.New()
	for _, sha := range keys {
		h.Write([]byte(sha))
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func DownloadFile(url, target string) error {
	file, err := os.Create(target)
	if err != nil {
		return err
	}
	defer file.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}
