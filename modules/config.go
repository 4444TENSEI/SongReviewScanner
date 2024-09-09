package modules

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
)

// 搜索选项的结构体
type SearchOptions struct {
	PlaylistIDs   []int64 `json:"playlists"`
	TrackIDs      []int64 `json:"singles"`
	ListeningRank struct {
		RecentWeekAllTime int `json:"week/all"`
	} `json:"listening"`
}

// 配置的结构体
type Config struct {
	TargetUserIDs        []int64       `json:"userId"`
	SearchMethod         string        `json:"searchMethod"`
	SearchOptions        SearchOptions `json:"searchOptions"`
	MaxPages             int           `json:"maxPages"`
	ConcurrentThreads    int           `json:"queriesPerSecond"`
	IgnoreCompletedTasks bool          `json:"ignoreCompletedTasks"`
}

// 函数读取配置文件并返回 Config 实例
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, errors.New("读取配置文件失败：" + err.Error())
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, errors.New("解析配置文件失败：" + err.Error())
	}

	// 动态验证 SearchMethod 是否有效
	optionsValue := reflect.ValueOf(cfg.SearchOptions)
	optionsType := optionsValue.Type()
	methodIsValid := false
	for i := 0; i < optionsType.NumField(); i++ {
		field := optionsType.Field(i)
		if cfg.SearchMethod == field.Tag.Get("json") {
			methodIsValid = true
			break
		}
	}

	if !methodIsValid {
		return nil, errors.New("search_method 必须是 search_options 中的一个键")
	}

	// 确保配置文件config.json中的week/all是0或1
	if cfg.SearchMethod == "listening" {
		listeningRank := optionsValue.FieldByName("ListeningRank").FieldByName("RecentWeekAllTime").Int()
		if listeningRank != 0 && listeningRank != 1 {
			return nil, errors.New("当search_method是listening时，recent_week/all_time 必须是0或1")
		}
	}

	return &cfg, nil
}

// 方法返回与搜索方法对应的搜索选项
func (c *Config) GetSearchOptionByMethod() interface{} {
	optionsValue := reflect.ValueOf(c.SearchOptions)
	optionsType := optionsValue.Type()

	for i := 0; i < optionsType.NumField(); i++ {
		field := optionsType.Field(i)
		if c.SearchMethod == field.Tag.Get("json") {
			return optionsValue.Field(i).Interface()
		}
	}

	return nil
}
