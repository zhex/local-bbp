package integrations

import (
	"encoding/json"
	"io"
	"net/http"
)

const dataUrl = "https://bitbucket.org/bitbucketpipelines/official-pipes/raw/master/pipes.prod.json"

func Search() ([]Integration, error) {
	var integrations []Integration

	resp, err := http.Get(dataUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(data, &integrations)
	if err != nil {
		return nil, err
	}
	return integrations, err
}
