
<p align="center"><img src="https://testingcf.jsdelivr.net/gh/4444TENSEI/CDN/img/avatar/AngelDog/AngelDog-rounded.png" alt="Logo"
    width="200" height="200"/></p>
<h1 align="center">SongReviewScanner</h1>
<h3 align="center">查找冈易云音乐指定用户的"歌曲评论"自动化脚本</h3>
<p align="center">
    <img src="https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
    <img src="https://img.shields.io/badge/json-5E5C5C?style=for-the-badge&logo=json&logoColor=white" />
</p>

<br/>

# 目录

- ### [部署](#部署-1)

- ### [修改配置](#修改配置文件-1)

  - ### [参数说明](#参数说明-1)

- ### [目录结构](#目录结构-1)

  

<hr/>

# 部署

### 拉取项目

```
git clone https://github.com/4444TENSEI/SongReviewScanner
```

### 启动

```
go run main.go
```



<hr/>

# 修改配置文件

### `config.json`示例:

```json
{
    "userId": [
        666666
    ],
    "searchMethod": "playlists",
    "searchOptions": {
        "playlists": [
            666666,888888
        ],
        "singles": [
            666666,888888
        ]
    },
    "maxPages": 666666,
    "queriesPerSecond": 500
}
```

# 参数说明

|       参数       | 值                                                           |
| :--------------: | :----------------------------------------------------------- |
|      userId      | `数值`，目标用户ID，从网页端个人首页地址栏最后的id=获取，https://music.163.com/#/user/home?id=1 |
|   searchMethod   | `字符串`，规定查找方式，`从下方searchOptions中选择`          |
|  searchOptions   | `playlists`和`singles`，分别是歌单方式和单曲方式，分别可以放入多个数值。至于listening不用管因为没做，包括week/all也不用管。 |
|     maxPages     | `数值`，上限为20000，对于每首歌曲的查找封顶数量限制。比如设置为999那么每首歌曲只会查999条。超出了20000也没用，因为接口会返回重复的响应。 |
| queriesPerSecond | `数值`，代表并发数，越大越快，比如一首歌单有1000首歌，我设置200或者更高甚至根据歌曲数目直接拉满. |



<hr/>

# 目录结构

```
SongReviewScanner
├─ go.mod
├─ main.go		//主程序
├─ script
│  ├─ build.bat		//打包脚本
├─ modules
│  ├─ config.go		//加载配置文件
│  ├─ playlists.go	//获取歌单内的单曲信息
│  └─ singles.go	//从主程序传入单曲数值数组，负责主要的爬取工作
└─ asset		//打包用的资源文件
```
