package main

import (
	"fmt"
	"log"

	modules "github.com/4444TENSEI/SongReviewScanner/modules"
)

func main() {
	cfg, err := modules.LoadConfig("config.json")
	if err != nil {
		log.Fatalf("加载配置文件失败：%v", err)
	}

	userIDsString := fmt.Sprintf("%v", cfg.TargetUserIDs)
	fmt.Printf("\n--------\n| 配置文件载入成功：\n| 目标用户ID：%s\n| 搜索方法：%s\n| 搜索选项：%v\n| 每首歌查询上限：%d\n| 并发限制：%d\n",
		userIDsString,
		cfg.SearchMethod,
		cfg.GetSearchOptionByMethod(),
		cfg.MaxPages,
		cfg.ConcurrentThreads,
	)

	// 获取歌曲ID切片
	var songIDs []string
	switch cfg.SearchMethod {
	case "playlists":
		playlistsInterface := cfg.GetSearchOptionByMethod()

		if playlistsInterface == nil {
			log.Fatalf("search_method %s 没有对应的搜索选项", cfg.SearchMethod)
		}

		playlists, ok := playlistsInterface.([]int64)
		if !ok {
			log.Fatalf("search_method %s 对应的搜索选项类型错误，期望 []int64，得到 %T", cfg.SearchMethod, playlistsInterface)
		}

		songIDsMap, err := modules.ScrapePlaylist(playlists)
		if err != nil {
			log.Fatalf("获取播放列表中的歌曲ID失败：%v", err)
		}

		for _, songIDsList := range songIDsMap {
			for _, songID := range songIDsList {
				songIDs = append(songIDs, fmt.Sprintf("%d", songID))
			}
		}
		// 在获取歌曲ID切片后，添加以下代码来输出ID数量
		fmt.Printf("| 歌单中的歌曲数量：%d\n--------\n", len(songIDs))

	case "singles":
		singlesInterface := cfg.GetSearchOptionByMethod()
		if singlesInterface == nil {
			log.Fatalf("search_method %s 没有对应的搜索选项", cfg.SearchMethod)
		}

		singles, ok := singlesInterface.([]int64)
		if !ok {
			log.Fatalf("search_method %s 对应的搜索选项类型错误，期望 []int64，得到 %T", cfg.SearchMethod, singlesInterface)
		}

		for _, single := range singles {
			songIDs = append(songIDs, fmt.Sprintf("%d", single))
		}

	default:
		log.Fatalf("未知的搜索方法：%s", cfg.SearchMethod)
	}

	// 抓取评论
	for _, userID := range cfg.TargetUserIDs {
		targetUserID := int(userID)
		if err := modules.FetchCommentsForSongs(songIDs, targetUserID, cfg.MaxPages, cfg.ConcurrentThreads); err != nil {
			log.Fatalf("抓取评论失败：%v", err)
		}
	}

	fmt.Println("评论抓取完成。")
}
