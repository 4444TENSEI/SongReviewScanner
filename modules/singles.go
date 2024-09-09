package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func getContents(songID string, page int, userID int) ([]Comment, int, error) {
	url := fmt.Sprintf("https://music.163.com/api/v1/resource/comments/R_SO_4_%s?limit=100&offset=%d", songID, page)
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, 0, err
	}

	// 检查 "total" 是否存在且不是 nil
	if total, ok := data["total"].(float64); ok {
		totalInt := int(total)
		var comments []Comment
		if items, ok := data["comments"].([]interface{}); ok {
			for _, item := range items {
				comment := item.(map[string]interface{})
				user := comment["user"].(map[string]interface{})
				commentUserID := int(user["userId"].(float64))
				if commentUserID != userID {
					continue
				}
				commentContent := comment["content"].(string)
				commentTime := comment["time"].(float64)
				commentTimestamp := time.Unix(int64(commentTime/1000), 0).Format("2006-01-02 15:04:05")
				commentLocation := ""
				if location, ok := comment["location"].(string); ok {
					commentLocation = location
				}

				commentData := Comment{
					Username:  user["nickname"].(string),
					Comment:   commentContent,
					SongID:    songID,
					Timestamp: commentTimestamp,
					Location:  commentLocation,
				}
				comments = append(comments, commentData)
			}
			return comments, totalInt, nil
		}
	}

	return nil, 0, fmt.Errorf("invalid response data")
}

func appendToJSONFile(comments []Comment, filename string) error {
	for _, comment := range comments {
		var fileData []Comment
		file, err := ioutil.ReadFile(filename)
		if err != nil {
			if os.IsNotExist(err) {
				fileData = []Comment{comment}
			} else {
				return err
			}
		} else {
			err = json.Unmarshal(file, &fileData)
			if err != nil {
				return err
			}
			fileData = append(fileData, comment)
		}
		newData, err := json.MarshalIndent(fileData, "", "    ")
		if err != nil {
			return err
		}
		err = ioutil.WriteFile(filename, newData, 0644)
		if err != nil {
			return err
		}
	}
	return nil
}

func worker(songID string, _ *sync.WaitGroup, ch chan error, filename string, userID int, cacheDir string, maxPages int, threadID int, semaphore chan struct{}) {
	<-semaphore             // 等待信号量，确保不超过并发限制
	startTime := time.Now() // 记录任务开始时间

	// 输出当前执行的歌曲ID和这首歌获取到的评论总数
	cache, err := LoadCache(songID, cacheDir, userID)

	if err != nil {
		if os.IsNotExist(err) {
			// 如果缓存不存在，则初始化缓存数据
			cache = CacheData{TotalPages: 0, FinalPage: 0}
		} else {
			ch <- err
			semaphore <- struct{}{}
			return
		}
	}

	// 检测缓存中是否存在该歌曲ID和用户ID的缓存，如果存在且缓存已完成则跳过
	if cache.TotalPages > 0 && cache.FinalPage >= cache.TotalPages {
		ch <- nil
		semaphore <- struct{}{}
		return
	}

	commentCount := 0
	var totalQueries int
	for i := cache.FinalPage; i < maxPages; i += 100 {
		totalQueries++
		comments, newTotal, err := getContents(songID, i, userID)
		if err != nil {
			ch <- err
			semaphore <- struct{}{}
			return
		}
		commentCount += len(comments)
		if newTotal > cache.TotalPages {
			cache.TotalPages = newTotal
			err = SaveCache(songID, cacheDir, cache, userID)
			if err != nil {
				ch <- err
				semaphore <- struct{}{}
				return
			}
		}
		err = appendToJSONFile(comments, filename)
		if err != nil {
			ch <- err
			semaphore <- struct{}{}
			return
		}
		cache.FinalPage = i + 100
		err = SaveCache(songID, cacheDir, cache, userID)
		if err != nil {
			ch <- err
			semaphore <- struct{}{}
			return
		}
		// 如果页面偏移量大于或等于评论总数，停止抓取
		if cache.FinalPage >= cache.TotalPages {
			break
		}
	}

	// 计算任务花费时间
	elapsedTime := time.Since(startTime).Seconds()

	if commentCount == 0 {
		fmt.Printf("××× 任务%d-歌曲[%s]任务完成，查询%d条评论, %.2f秒，没有找到目标评论。\n", threadID, songID, totalQueries*100, elapsedTime)
	} else {
		fmt.Printf("√√√ 任务%d-歌曲[%s]任务完成，查询%d条评论, %.2f秒，找到了%d个目标评论!!! \n", threadID, songID, totalQueries*100, elapsedTime, commentCount)
	}
	ch <- nil
	semaphore <- struct{}{}
}

// 函数接受歌曲ID数组、用户ID、最大页数、并发线程数，并开始抓取评论
func FetchCommentsForSongs(songIDs []string, userID int, maxPages int, queriesPerSecond int) error {
	outputDir := "output"
	cacheDir := "cache"

	os.MkdirAll(outputDir, 0755)
	os.MkdirAll(cacheDir, 0755)

	filename := filepath.Join(outputDir, fmt.Sprintf("user_%d.json", userID))

	var wg sync.WaitGroup
	errCh := make(chan error, len(songIDs))
	semaphore := make(chan struct{}, queriesPerSecond)
	tasks := make(chan string, len(songIDs))

	// 启动一个goroutine来分发任务
	go func() {
		for _, songID := range songIDs {
			tasks <- songID // 将任务放入队列
		}
		close(tasks)
	}()

	// 启动goroutines来处理任务
	for i := 0; i < queriesPerSecond; i++ {
		wg.Add(1)
		go func(threadID int) {
			defer wg.Done()
			for songID := range tasks { // 从任务队列中获取任务
				semaphore <- struct{}{}
				ch := make(chan error)
				go worker(songID, &wg, ch, filename, userID, cacheDir, maxPages, threadID, semaphore)
				if err := <-ch; err != nil {
					errCh <- err
				}
				<-semaphore
			}
		}(i + 1)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}
