package afdian

import (
	"AfdianToMarkdown/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/carlmjohnson/requests"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var Host string

const (
	DelayMs         = 330
	ChromeUserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
)

type Album struct {
	AlbumName string `json:"albumName"`
	AlbumUrl  string `json:"albumUrl"`
}

type Article struct {
	ArticleName string `json:"articleName"`
	ArticleUrl  string `json:"articleUrl"`
}

// Cookie 从 Chrome 中使用cookie master导出的 Cookies
type Cookie struct {
	Domain     string  `json:"domain"`
	Expiration float64 `json:"expirationDate"`
	HostOnly   bool    `json:"hostOnly"`
	HTTPOnly   bool    `json:"httpOnly"`
	Name       string  `json:"name"`
	Path       string  `json:"path"`
	SameSite   string  `json:"sameSite"`
	Secure     bool    `json:"secure"`
	Session    bool    `json:"session"`
	StoreID    string  `json:"storeId"`
	Value      string  `json:"value"`
}

// ReadCookiesFromFile 从文件中读取 Cookies
func ReadCookiesFromFile(filePath string) []Cookie {
	var cookies []Cookie

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(data, &cookies)
	if err != nil {
		log.Fatal(err)
	}

	return cookies
}

func GetCookiesString(cookies []Cookie) string {
	var cookieString string
	for _, cookie := range cookies {
		cookieString += cookie.Name + "=" + cookie.Value + ";"
	}
	return cookieString
}

func GetAuthTokenCookieString(cookies []Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			return fmt.Sprintf("auth_token=%s", cookie.Value)
		}
	}
	return ""
}

func buildAfdianHeaders(cookieString string, referer string) http.Header {
	return http.Header{
		"authority":          {"afdian.net"},
		"accept":             {"accept", "application/json, text/plain, */*"},
		"accept-language":    {"zh-CN,zh;q=0.9,en;q=0.8"},
		"afd-fe-version":     {"20220508"},
		"afd-stat-id":        {"c78521949a7c11ee8c2452540025c377"},
		"cache-control":      {"no-cache"},
		"cookie":             {cookieString},
		"dnt":                {"1"},
		"locale-lang":        {"zh-CN"},
		"pragma":             {"no-cache"},
		"referer":            {referer},
		"sec-ch-ua":          {`"Chromium";v="127", "Not(A:Brand";v="99", "Google Chrome";v="127"`},
		"sec-ch-ua-mobile":   {"?0"},
		"sec-ch-ua-platform": {`"Windows"`},
		"sec-fetch-dest":     {"empty"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-site":     {"same-origin"},
		"sec-gpc":            {"1"},
		"user-agent":         {ChromeUserAgent},
	}
}

// NewRequestGet 发送GET请求
func NewRequestGet(Url string, cookieString string, referer string) []byte {
	var body bytes.Buffer
	err := requests.
		URL(Url).
		Headers(buildAfdianHeaders(cookieString, referer)).
		ToBytesBuffer(&body).
		Fetch(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	return body.Bytes()
}

// GetAuthorId 获取作者的ID
// refer: https://afdian.net/a/xyzName
func GetAuthorId(authorName string, referer string, cookieString string) string {
	apiUrl := fmt.Sprintf("%s/api/user/get-profile-by-slug?url_slug=%s", Host, authorName)
	body := NewRequestGet(apiUrl, cookieString, referer)
	//fmt.Printf("%s\n", body)
	authorId := gjson.GetBytes(body, "data.user.user_id").String()
	return authorId
}

// GetAuthorArticleUrlListByInterface 获取作者的文章列表
// publish_sn获取的逻辑是第一轮请求为空，然后第二轮请求输入上一轮获取到的最后一篇文章的publish_sn，以此类推，直到获取到的publish_sn为空结束
func GetAuthorArticleUrlListByInterface(userName string, cookieString string, prevPublishSn string) ([]Article, string) {
	userReferer := fmt.Sprintf("%s/a/%s", Host, userName)
	userId := GetAuthorId(userName, userReferer, cookieString)
	apiUrl := fmt.Sprintf("%s/api/post/get-list?user_id=%s&type=new&publish_sn=%s&per_page=10&group_id=&all=1&is_public=&plan_id=&title=&name=", Host, userId, prevPublishSn)
	log.Println("Get publish_sn apiUrl:", apiUrl)
	var authorArticleList []Article

	body := NewRequestGet(apiUrl, cookieString, userReferer)
	//log.Printf("%s\n", body)

	articleListJson := gjson.GetBytes(body, "data.list")
	articleListJson.ForEach(func(key, value gjson.Result) bool {
		articleId := value.Get("post_id").String()
		articleUrl, _ := url.JoinPath(Host, "post", articleId)
		articleName := value.Get("title").String()
		authorArticleList = append(authorArticleList, Article{ArticleName: utils.ToSafeFilename(articleName), ArticleUrl: articleUrl})
		return true
	})

	publishSn := gjson.GetBytes(body, fmt.Sprintf("data.list.%d.publish_sn", len(authorArticleList)-1)).String()
	log.Println("publishSn:", publishSn)
	return authorArticleList, publishSn
}

// GetAlbumListByInterface 获取作者的作品集列表
func GetAlbumListByInterface(userId string, referer string, cookieString string) []Album {
	apiUrl := fmt.Sprintf("%s/api/user/get-album-list?user_id=%s", Host, userId)
	body := NewRequestGet(apiUrl, cookieString, referer)
	//fmt.Printf("%s\n", body)
	var albumList []Album
	albumListJson := gjson.GetBytes(body, "data.list")
	//fmt.Println(utils.ToJSON(albumListJson))
	albumListJson.ForEach(func(key, value gjson.Result) bool {
		//fmt.Println(value.Get("title").String())
		//fmt.Println(value.Get("album_id").String())
		albumId := value.Get("album_id").String()
		albumUrl, _ := url.JoinPath(Host, "album", albumId)
		albumList = append(albumList, Album{AlbumName: value.Get("title").String(), AlbumUrl: albumUrl})
		return true
	})
	//fmt.Println(albumList)
	return albumList
}

// GetAlbumArticleListByInterface 获取作品集的所有文章
func GetAlbumArticleListByInterface(albumId string, authToken string) []Article {
	//log.Println("albumId:", albumId)
	postCountApiUrl := fmt.Sprintf("%s/api/user/get-album-info?album_id=%s", Host, albumId)
	authTokenCookie := fmt.Sprintf("auth_token=%s", authToken)
	referer := fmt.Sprintf("%s/album/%s", Host, albumId)

	postCountBodyText := NewRequestGet(postCountApiUrl, authTokenCookie, referer)
	postCount := gjson.GetBytes(postCountBodyText, "data.album.post_count").Int()
	//log.Println("postCount:", postCount)

	var albumArticleList []Article
	var i int64
	for i = 0; i < postCount; i += 10 {
		apiUrl := fmt.Sprintf("%s/api/user/get-album-post?album_id=%s&lastRank=%d&rankOrder=asc&rankField=rank", Host, albumId, i)
		body := NewRequestGet(apiUrl, authTokenCookie, referer)

		albumArticleListJson := gjson.GetBytes(body, "data.list")
		albumArticleListJson.ForEach(func(key, value gjson.Result) bool {
			//fmt.Println(value.Get("title").String())
			//fmt.Println(value.Get("post_id").String())
			postId := value.Get("post_id").String()
			postUrl, _ := url.JoinPath(Host, "album", albumId, postId)
			albumArticleList = append(albumArticleList, Article{ArticleName: utils.ToSafeFilename(value.Get("title").String()), ArticleUrl: postUrl})
			return true
		})
	}

	return albumArticleList
}

// GetArticleContentByInterface 获取文章正文内容
func GetArticleContentByInterface(articleUrl string, authToken string, converter *md.Converter) string {
	//在album内的： https://afdian.net/api/post/get-detail?post_id={post_id}&album_id={album_id}
	//在album外的： https://afdian.net/api/post/get-detail?post_id={post_id}&album_id=
	log.Println("articleUrl:", articleUrl)
	var apiUrl string
	splitUrl := strings.Split(articleUrl, "/")
	if strings.Contains(articleUrl, "album") {
		apiUrl = fmt.Sprintf("%s/api/post/get-detail?post_id=%s&album_id=%s", Host, splitUrl[len(splitUrl)-1], splitUrl[len(splitUrl)-2])
	} else {
		apiUrl = fmt.Sprintf("%s/api/post/get-detail?post_id=%s&album_id=", Host, splitUrl[len(splitUrl)-1])
	}
	log.Println("Get article content apiUrl:", apiUrl)
	body := NewRequestGet(apiUrl, authToken, articleUrl)
	//log.Println("body: ", string(body))
	articleContent := gjson.GetBytes(body, "data.post.content").String()
	//log.Println("articleContent: ", articleContent)

	markdown, err := converter.ConvertString(articleContent)
	if err != nil {
		log.Fatal(err)
	}
	//log.Println(markdown)

	return markdown
}

// GetArticleCommentByInterface 获取文章评论
// TODO:根据publish_sn获取全部评论
// https://afdian.net/api/comment/get-list?post_id={post_id}&publish_sn={publish_sn}&type=old&hot=
func GetArticleCommentByInterface(articleUrl string, cookieString string) (commentString string, hotCommentString string) {
	//https://afdian.net/api/comment/get-list?post_id={post_id}&publish_sn=&type=old&hot=1
	splitUrl := strings.Split(articleUrl, "/")
	postId := splitUrl[len(splitUrl)-1]
	apiUrl := fmt.Sprintf("%s/api/comment/get-list?post_id=%s&publish_sn=&type=old&hot=1", Host, postId)
	log.Println("Get article comment apiUrl:", apiUrl)

	body := NewRequestGet(apiUrl, cookieString, articleUrl)
	commentJson := gjson.GetBytes(body, "data.list")
	hotCommentJson := gjson.GetBytes(body, "data.hot_list")
	if hotCommentJson.Exists() {
		hotCommentString += "## 热评\n\n" + getCommentString(hotCommentJson)
	}

	commentString += "## 评论\n\n" + getCommentString(commentJson)

	return commentString, hotCommentString
}

func getCommentString(commentJson gjson.Result) string {
	i := 0
	hotCommentString := ""
	commentJson.ForEach(func(key, value gjson.Result) bool {
		nickName := value.Get("user.name").String()
		publishTimeStamp := cast.ToInt64(value.Get("publish_time").String())
		publishTime := time.Unix(publishTimeStamp, 0).Format("2006-01-02 15:04:05")
		content := value.Get("content").String()
		replyString := ""
		replyUser := value.Get("reply_user")
		if replyUser.Exists() {
			replyUserNickName := replyUser.Get("name").String()
			replyString = fmt.Sprintf("> 回复 %s: ", replyUserNickName)
		}
		hotCommentString += fmt.Sprintf("----\n##### <span>[%d] %s by %s</span>\n%s\n\n%s\n\n", i, publishTime, nickName, replyString, content)
		//fmt.Println(hotCommentString)
		i++
		return true
	})
	return hotCommentString
}

func SaveContentIfNotExist(articleName string, filePath string, articleUrl string, authToken string, converter *md.Converter) error {
	_, fileExists := utils.FileExists(filePath)
	log.Println("fileExists:", fileExists)
	//如果文件不存在，则下载
	if !fileExists {
		content := GetArticleContentByInterface(articleUrl, authToken, converter)
		commentString, hotCommentString := GetArticleCommentByInterface(articleUrl, authToken)
		//Refer中需要把articleUrl中的post替换成p才能在浏览器正常访问
		articleContent := "## " + articleName + "\n\n### Refer\n\n" + strings.Replace(articleUrl, "post", "p", 1) + "\n\n### 正文\n\n" + fmt.Sprintf("%s\n\n%s\n\n%s", content, hotCommentString, commentString)
		//log.Println("articleContent:", articleContent)
		if err := os.WriteFile(filePath, []byte(articleContent), os.ModePerm); err != nil {
			return err
		}
		time.Sleep(time.Millisecond * time.Duration(DelayMs))
	} else {
		log.Println(filePath, "已存在，跳过下载")
	}
	return nil
}
