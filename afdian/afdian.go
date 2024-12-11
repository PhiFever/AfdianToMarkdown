package afdian

import (
	"AfdianToMarkdown/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/carlmjohnson/requests"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
)

var (
	Host string
)

const (
	DelayMs         = 150
	ChromeUserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
)

type Album struct {
	AlbumName string
	AlbumUrl  string
}

type AlbumPost interface {
	GetName() string
	GetUrl() string
}

type Article struct {
	Name string
	Url  string
}

func (a Article) GetName() string {
	return a.Name
}

func (a Article) GetUrl() string {
	return a.Url
}

type Manga struct {
	Name     string
	Url      string
	Pictures []string
}

func (m Manga) GetName() string {
	return m.Name
}

func (m Manga) GetUrl() string {
	return m.Url
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

func SetHost(afdianHost string) {
	Host = fmt.Sprintf("https://%s", afdianHost)
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

func GetCookiesString(cookies []Cookie) (cookiesString string) {
	for _, cookie := range cookies {
		cookiesString += cookie.Name + "=" + cookie.Value + ";"
	}
	return cookiesString
}

func GetAuthTokenString(cookies []Cookie) (authTokenString string) {
	for _, cookie := range cookies {
		if cookie.Name == "auth_token" {
			authTokenString = fmt.Sprintf("auth_token=%s", cookie.Value)
		}
	}
	return authTokenString
}

func GetCookies() (cookieString string, authToken string) {
	cookies := ReadCookiesFromFile(utils.CookiePath)
	cookieString = GetCookiesString(cookies)
	//log.Println("cookieString:", cookieString)
	authToken = GetAuthTokenString(cookies)
	return cookieString, authToken
}

func buildAfdianHeaders(cookieString string, referer string) http.Header {
	return http.Header{
		"authority":          {"afdian.com"},
		"accept":             {"accept", "application/json, text/plain, */*"},
		"accept-language":    {"zh-CN,zh;q=0.9,en;q=0.8"},
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
// refer: https://afdian.com/a/Alice
func GetAuthorId(authorName string, referer string, cookieString string) (authorId string) {
	apiUrl := fmt.Sprintf("%s/api/user/get-profile-by-slug?url_slug=%s", Host, authorName)
	body := NewRequestGet(apiUrl, cookieString, referer)
	//fmt.Printf("%s\n", body)
	authorId = gjson.GetBytes(body, "data.user.user_id").String()
	return authorId
}

// GetAuthorMotionUrlList 获取作者的文章列表
// publish_sn获取的逻辑是第一轮请求为空，然后第二轮请求输入上一轮获取到的最后一篇文章的publish_sn，以此类推，直到获取到的publish_sn为空结束
func GetAuthorMotionUrlList(userName string, cookieString string, prevPublishSn string) (authorArticleList []Article, nextPublishSn string) {
	userReferer := fmt.Sprintf("%s/a/%s", Host, userName)
	userId := GetAuthorId(userName, userReferer, cookieString)
	apiUrl := fmt.Sprintf("%s/api/post/get-list?user_id=%s&type=new&publish_sn=%s&per_page=10&group_id=&all=1&is_public=&plan_id=&title=&name=", Host, userId, prevPublishSn)
	log.Println("Get publish_sn apiUrl:", apiUrl)

	body := NewRequestGet(apiUrl, cookieString, userReferer)
	//log.Printf("%s\n", body)

	articleListJson := gjson.GetBytes(body, "data.list")
	articleListJson.ForEach(func(key, value gjson.Result) bool {
		articleId := value.Get("post_id").String()
		articleUrl, _ := url.JoinPath(Host, "post", articleId)
		articleName := value.Get("title").String()
		authorArticleList = append(authorArticleList, Article{Name: utils.ToSafeFilename(articleName), Url: articleUrl})
		return true
	})

	nextPublishSn = gjson.GetBytes(body, fmt.Sprintf("data.list.%d.publish_sn", len(authorArticleList)-1)).String()
	log.Println("nextPublishSn:", nextPublishSn)
	return authorArticleList, nextPublishSn
}

// GetAlbumList 获取作者的作品集列表
func GetAlbumList(userId string, referer string, cookieString string) (albumList []Album) {
	apiUrl := fmt.Sprintf("%s/api/user/get-album-list?user_id=%s", Host, userId)
	body := NewRequestGet(apiUrl, cookieString, referer)
	//fmt.Printf("%s\n", body)
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

func GetAlbumPostList(albumId string, cookieString string) (albumPostList []AlbumPost) {
	postCountApiUrl := fmt.Sprintf("%s/api/user/get-album-info?album_id=%s", Host, albumId)
	referer := fmt.Sprintf("%s/album/%s", Host, albumId)

	postCountBodyText := NewRequestGet(postCountApiUrl, cookieString, referer)
	postCount := gjson.GetBytes(postCountBodyText, "data.album.post_count").Int()

	var i int64
	for i = 0; i < postCount; i += 10 {
		apiUrl := fmt.Sprintf("%s/api/user/get-album-post?album_id=%s&lastRank=%d&rankOrder=asc&rankField=rank", Host, albumId, i)
		body := NewRequestGet(apiUrl, cookieString, referer)

		albumPostListJson := gjson.GetBytes(body, "data.list")
		albumPostListJson.ForEach(func(key, value gjson.Result) bool {
			postId := value.Get("post_id").String()
			postUrl, _ := url.JoinPath(Host, "album", albumId, postId)

			if value.Get("pics").Exists() {
				// 如果有图片，则创建一个 Manga 对象
				var pictures []string
				for _, result := range value.Get("pics").Array() {
					pictures = append(pictures, result.String())
				}
				albumPostList = append(albumPostList, Manga{
					Name:     utils.ToSafeFilename(value.Get("title").String()),
					Url:      postUrl,
					Pictures: pictures,
				})
			} else {
				// 否则创建一个 Article 对象
				albumPostList = append(albumPostList, Article{
					Name: utils.ToSafeFilename(value.Get("title").String()),
					Url:  postUrl,
				})
			}
			return true
		})
	}

	return albumPostList
}

// GetArticleContent 获取文章正文内容
func GetArticleContent(articleUrl string, authToken string, converter *md.Converter) string {
	//在album内的： https://afdian.com/api/post/get-detail?post_id={post_id}&album_id={album_id}
	//在album外的： https://afdian.com/api/post/get-detail?post_id={post_id}&album_id=
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

// GetArticleComment 获取文章评论
// TODO:根据publish_sn获取全部评论
// https://afdian.com/api/comment/get-list?post_id={post_id}&publish_sn={publish_sn}&type=old&hot=
func GetArticleComment(articleUrl string, cookieString string) (commentString string, hotCommentString string) {
	//https://afdian.com/api/comment/get-list?post_id={post_id}&publish_sn=&type=old&hot=1
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

func getCommentString(commentJson gjson.Result) (commentString string) {
	i := 0
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
		commentString += fmt.Sprintf("----\n##### <span>[%d] %s by %s</span>\n%s\n\n%s\n\n", i, publishTime, nickName, replyString, content)
		//fmt.Println(commentString)
		i++
		return true
	})
	return commentString
}

func SaveContentIfNotExist(articleName string, filePath string, articleUrl string, authToken string, converter *md.Converter) error {
	_, err := os.Stat(filePath)
	fileExists := err == nil || os.IsExist(err)
	log.Println("fileExists:", fileExists)
	//如果文件不存在，则下载
	if !fileExists {
		content := GetArticleContent(articleUrl, authToken, converter)
		commentString, hotCommentString := GetArticleComment(articleUrl, authToken)
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

func SaveMangaIfNotExist(filePath string, manga Manga, authToken string, converter *md.Converter) error {
	_, err := os.Stat(filePath)
	fileExists := err == nil || os.IsExist(err)
	log.Println("Picture Exists:", fileExists)

	if !fileExists {
		assetsDir := filepath.Join(filepath.Dir(filePath), utils.ImgDir)
		if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
			return fmt.Errorf("create assets directory error: %v", err)
		}
		content := GetArticleContent(manga.Url, authToken, converter)
		picContent := ""
		// 下载并保存图片到本地
		for i, pictureUrl := range manga.Pictures {
			// 生成本地图片文件名
			ext := filepath.Ext(pictureUrl)
			if ext == "" {
				ext = ".jpg" // 默认扩展名
			}
			localFileName := fmt.Sprintf("%s_%d%s", utils.ToSafeFilename(manga.Name), i, ext)
			localFilePath := filepath.Join(assetsDir, localFileName)

			log.Printf("Downloading picture in manga %s: %s", manga.Name, pictureUrl)
			// 使用requests下载图片
			err := requests.
				URL(pictureUrl).
				//Header("Authorization", fmt.Sprintf("Bearer %s", authToken)).
				ToFile(localFilePath).
				Fetch(context.Background())

			if err != nil {
				log.Printf("Failed to download image %s: %v", pictureUrl, err)
				// 如果下载失败，使用原始URL
				picContent += fmt.Sprintf("![image](%s)\n", pictureUrl)
				continue
			}

			// 使用相对路径引用本地图片
			relPath := filepath.Join(utils.ImgDir, localFileName)
			picContent += fmt.Sprintf("![image](%s)\n", relPath)
		}

		commentString, hotCommentString := GetArticleComment(manga.Url, authToken)
		//Refer中需要把articleUrl中的post替换成p才能在浏览器正常访问
		articleContent := "## " + manga.Name + "\n\n### Refer\n\n" + strings.Replace(manga.Url, "post", "p", 1) + "\n\n### 正文\n\n" + fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", content, picContent, hotCommentString, commentString)
		//log.Println("articleContent:", articleContent)
		if err := os.WriteFile(filePath, []byte(articleContent), os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
