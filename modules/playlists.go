package modules

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// 根据歌单，所有歌曲id
func ScrapePlaylist(playlistIDs []int64) (map[int64][]int64, error) {
	songIDs := make(map[int64][]int64)

	for _, id := range playlistIDs {
		url := fmt.Sprintf("https://music.163.com/api/v6/playlist/detail?id=%d", id)

		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("请求歌单API失败：%v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("请求歌单API失败，状态码：%d", resp.StatusCode)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应体失败：%v", err)
		}

		var playlistDetail struct {
			Code     int `json:"code"`
			Playlist struct {
				TrackIds []struct {
					ID int64 `json:"id"`
				} `json:"trackIds"`
			} `json:"playlist"`
		}

		err = json.Unmarshal(body, &playlistDetail)
		if err != nil {
			return nil, fmt.Errorf("解析JSON数据失败：%v", err)
		}

		if playlistDetail.Code != 200 {
			return nil, fmt.Errorf("API返回错误代码：%d", playlistDetail.Code)
		}

		var songIDList []int64
		for _, trackID := range playlistDetail.Playlist.TrackIds {
			songIDList = append(songIDList, trackID.ID)
		}

		songIDs[id] = songIDList
	}

	return songIDs, nil
}
