package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type CacheData struct {
	TotalPages int `json:"totalPages"`
	FinalPage  int `json:"finalPage"`
}

type Comment struct {
	Username  string `json:"username"`
	Comment   string `json:"comment"`
	SongID    string `json:"song_id"`
	Location  string `json:"location"`
	Timestamp string `json:"time"`
}

// 创建输出目录和缓存目录
func CreateOutputDirs(outputDir, cacheDir string) {
	os.MkdirAll(outputDir, 0755)
	os.MkdirAll(cacheDir, 0755)
}

// 生成缓存文件名
func generateCacheFilename(songID string, userID int) string {
	return fmt.Sprintf("cache_%d_%s.json", userID, songID)
}

// 保存缓存数据
func SaveCache(songID string, cacheDir string, cache CacheData, userID int) error {
	progressFilename := generateCacheFilename(songID, userID)
	progressFile := filepath.Join(cacheDir, progressFilename)
	progressData, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(progressFile, progressData, 0644)
}

// 加载缓存数据
func LoadCache(songID string, cacheDir string, userID int) (CacheData, error) {
	progressFilename := generateCacheFilename(songID, userID)
	progressFile := filepath.Join(cacheDir, progressFilename)
	progressData, err := ioutil.ReadFile(progressFile)
	if err != nil {
		if os.IsNotExist(err) {
			return CacheData{}, err
		}
		return CacheData{}, err
	}
	var cache CacheData
	err = json.Unmarshal(progressData, &cache)
	if err != nil {
		return CacheData{}, err
	}
	return cache, nil
}
