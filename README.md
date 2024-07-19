## AifadianCrawler

爱发电(afdian.net)爬虫，用于下载爱发电作者的动态和作品集并保存为markdown文件。

## 准备

使用浏览器插件（如Edit this cookie,cookie master等）获取cookie，将cookie保存到项目根目录（即与exe可执行文件同级目录）的`cookie.json`文件中。

注意主站域名需要确认（默认为afdian.net)

## 使用方法

### 下载作者主页的所有动态

```shell
go run main.go --host="afdian.com" -au "作者id（主页url的最后一部分，如https://afdian.net/a/作者id/）" motions
```

### 下载作者所有的作品集

```shell
go run main.go -au "作者id" albums
```

### 下载指定文件中按行分隔的作品集（尚未实现）

```shell
go run main.go -l "文件路径"
```

## Update
## v0.2
由于主站(afdian.net)在7月15日被屏蔽，添加了对于手动更改临时域名(如afdian.com)的支持