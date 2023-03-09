package github

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Release API response 结构
type Release struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name string `json:"name"`
		Url  string `json:"browser_download_url"`
	} `json:"assets"`
}

// getApi 获取API地址
func getApi(repo string) string {
	return fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
}

// GetLatest 获取最新版本信息
func GetLatest(repo string) (Release, error) {
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
