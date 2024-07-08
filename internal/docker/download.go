package docker

import (
	"fmt"
	"github.com/zhex/local-bbp/internal/common"
	"os"
	"path"
	"sync"
)

var SupportedArchitectures = []string{"x86_64", "aarch64"}

func DownloadDockerCliBinary(version, target string) error {
	var wg sync.WaitGroup
	var errs []error

	for _, arch := range SupportedArchitectures {
		wg.Add(1)
		go func(archType string) {
			defer wg.Done()

			url := fmt.Sprintf("https://download.docker.com/linux/static/stable/%s/docker-%s.tgz", arch, version)
			tmpPath := fmt.Sprintf("/tmp/docker-%s-%s.tgz", version, archType)
			if err := common.DownloadFile(url, tmpPath); err != nil {
				errs = append(errs, err)
				return
			}
			file, err := os.Open(tmpPath)
			if err != nil {
				errs = append(errs, err)
				return
			}
			defer file.Close()

			finalTarget := path.Join(target, archType)
			if err := common.ExtractTarGz(file, finalTarget); err != nil {
				errs = append(errs, err)
			}
		}(arch)
	}

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("failed to download docker cli binary: %v", errs[0])
	}
	return nil
}
