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
	host    string
	HostUrl string
)

const (
	DelayMs         = 150
	ChromeUserAgent = `Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36`
)

type Album struct {
	AlbumName string
	AlbumUrl  string
}

type Post struct {
	Name     string
	Url      string
	Pictures []string
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

func SetHostUrl(afdianHost string) {
	host = afdianHost
	HostUrl = fmt.Sprintf("https://%s", afdianHost)
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
		"authority":          {host},
		"accept":             {"accept", "application/json, text/plain, */*"},
		"accept-language":    {"zh-CN,zh;q=0.9,en;q=0.8"},
		"cache-control":      {"no-cache"},
		"cookie":             {cookieString},
		"dnt":                {"1"},
		"locale-lang":        {"zh-CN"},
		"pragma":             {"no-cache"},
		"referer":            {referer},
		"sec-ch-ua":          {`"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`},
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
func GetAuthorId(authorUrlSlug string, referer string, cookieString string) (authorId string) {
	apiUrl := fmt.Sprintf("%s/api/user/get-profile-by-slug?url_slug=%s", HostUrl, authorUrlSlug)
	body := NewRequestGet(apiUrl, cookieString, referer)
	//fmt.Printf("%s\n", body)
	authorId = gjson.GetBytes(body, "data.user.user_id").String()
	return authorId
}

// GetMotionUrlList 获取作者的文章列表
// publish_sn获取的逻辑是第一轮请求为空，然后第二轮请求输入上一轮获取到的最后一篇文章的publish_sn，以此类推，直到获取到的publish_sn为空结束
func GetMotionUrlList(userName string, cookieString string, prevPublishSn string) (authorArticleList []Post, nextPublishSn string) {
	userReferer := fmt.Sprintf("%s/a/%s", HostUrl, userName)
	userId := GetAuthorId(userName, userReferer, cookieString)
	apiUrl := fmt.Sprintf("%s/api/post/get-list?user_id=%s&type=new&publish_sn=%s&per_page=10&group_id=&all=1&is_public=&plan_id=&title=&name=", HostUrl, userId, prevPublishSn)
	log.Println("Get publish_sn apiUrl:", apiUrl)

	body := NewRequestGet(apiUrl, cookieString, userReferer)
	//log.Printf("%s\n", body)

	articleListJson := gjson.GetBytes(body, "data.list")
	articleListJson.ForEach(func(key, value gjson.Result) bool {
		articleId := value.Get("post_id").String()
		articleUrl, _ := url.JoinPath(HostUrl, "post", articleId)
		var pictures []string
		for _, result := range value.Get("pics").Array() {
			pictures = append(pictures, result.String())
		}
		authorArticleList = append(authorArticleList, Post{
			Name:     utils.ToSafeFilename(value.Get("title").String()),
			Url:      articleUrl,
			Pictures: pictures,
		})
		return true
	})

	nextPublishSn = gjson.GetBytes(body, fmt.Sprintf("data.list.%d.publish_sn", len(authorArticleList)-1)).String()
	log.Println("nextPublishSn:", nextPublishSn)
	return authorArticleList, nextPublishSn
}

// GetAlbumList 获取作者的作品集列表
func GetAlbumList(userId string, referer string, cookieString string) (albumList []Album) {
	apiUrl := fmt.Sprintf("%s/api/user/get-album-list?user_id=%s", HostUrl, userId)
	body := NewRequestGet(apiUrl, cookieString, referer)
	//fmt.Printf("%s\n", body)
	albumListJson := gjson.GetBytes(body, "data.list")
	//fmt.Println(utils.ToJSON(albumListJson))
	albumListJson.ForEach(func(key, value gjson.Result) bool {
		//fmt.Println(value.Get("title").String())
		//fmt.Println(value.Get("album_id").String())
		albumId := value.Get("album_id").String()
		albumUrl, _ := url.JoinPath(HostUrl, "album", albumId)

		albumList = append(albumList, Album{AlbumName: value.Get("title").String(), AlbumUrl: albumUrl})
		return true
	})
	//fmt.Println(albumList)
	return albumList
}

func GetAlbumPostList(albumId string, cookieString string) (authorUrlSlug string, albumName string, albumPostList []Post) {
	postCountApiUrl := fmt.Sprintf("%s/api/user/get-album-info?album_id=%s", HostUrl, albumId)
	referer := fmt.Sprintf("%s/album/%s", HostUrl, albumId)

	postCountBodyText := NewRequestGet(postCountApiUrl, cookieString, referer)
	albumName = gjson.GetBytes(postCountBodyText, "data.album.title").String()
	postCount := gjson.GetBytes(postCountBodyText, "data.album.post_count").Int()
	authorUrlSlug = gjson.GetBytes(postCountBodyText, "data.album.user.url_slug").String()

	var i int64
	for i = 0; i < postCount; i += 10 {
		apiUrl := fmt.Sprintf("%s/api/user/get-album-post?album_id=%s&lastRank=%d&rankOrder=asc&rankField=rank", HostUrl, albumId, i)
		body := NewRequestGet(apiUrl, cookieString, referer)

		albumPostListJson := gjson.GetBytes(body, "data.list")
		albumPostListJson.ForEach(func(key, value gjson.Result) bool {
			postId := value.Get("post_id").String()
			postUrl, _ := url.JoinPath(HostUrl, "album", albumId, postId)

			var pictures []string
			for _, result := range value.Get("pics").Array() {
				pictures = append(pictures, result.String())
			}
			albumPostList = append(albumPostList, Post{
				Name:     utils.ToSafeFilename(value.Get("title").String()),
				Url:      postUrl,
				Pictures: pictures,
			})
			return true
		})
	}

	return authorUrlSlug, albumName, albumPostList
}

// getPostContent 获取文章正文内容
func getPostContent(articleUrl string, authToken string, converter *md.Converter) string {
	//在album内的： https://afdian.com/api/post/get-detail?post_id={post_id}&album_id={album_id}
	//在album外的： https://afdian.com/api/post/get-detail?post_id={post_id}&album_id=
	log.Println("articleUrl:", articleUrl)
	var apiUrl string
	splitUrl := strings.Split(articleUrl, "/")
	if strings.Contains(articleUrl, "album") {
		apiUrl = fmt.Sprintf("%s/api/post/get-detail?post_id=%s&album_id=%s", HostUrl, splitUrl[len(splitUrl)-1], splitUrl[len(splitUrl)-2])
	} else {
		apiUrl = fmt.Sprintf("%s/api/post/get-detail?post_id=%s&album_id=", HostUrl, splitUrl[len(splitUrl)-1])
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

// GetPostComment 获取文章评论
// TODO:根据publish_sn获取全部评论
// https://afdian.com/api/comment/get-list?post_id={post_id}&publish_sn={publish_sn}&type=old&hot=
func GetPostComment(articleUrl string, cookieString string) (commentString string, hotCommentString string) {
	//https://afdian.com/api/comment/get-list?post_id={post_id}&publish_sn=&type=old&hot=1
	splitUrl := strings.Split(articleUrl, "/")
	postId := splitUrl[len(splitUrl)-1]
	apiUrl := fmt.Sprintf("%s/api/comment/get-list?post_id=%s&publish_sn=&type=old&hot=1", HostUrl, postId)
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

func SavePostIfNotExist(filePath string, article Post, authToken string, disableComment bool, converter *md.Converter) error {
	_, err := os.Stat(filePath)
	fileExists := err == nil || os.IsExist(err)
	if !fileExists {
		log.Println("Saving file:", filePath)
		content := getPostContent(article.Url, authToken, converter)
		//TODO:不支持图文混排
		picContent, err := getPictures(filePath, article)
		if err != nil {
			return err
		}

		referUrl := strings.Replace(article.Url, "post", "p", 1)
		articleContent := fmt.Sprintf("## %s\n\n### Refer\n\n%s\n\n### 正文\n\n%s\n\n%s",
			article.Name, referUrl, content, picContent)

		if !disableComment {
			commentString, hotCommentString := GetPostComment(article.Url, authToken)
			articleContent = fmt.Sprintf("%s\n\n%s\n\n%s", articleContent, hotCommentString, commentString)
		}

		if err := os.WriteFile(filePath, []byte(articleContent), os.ModePerm); err != nil {
			return err
		}
	} else {
		log.Printf("File exists: %s", filePath)
	}
	return nil
}

func getPictures(filePath string, article Post) (string, error) {
	if len(article.Pictures) == 0 {
		return "", nil
	}
	assetsDir := filepath.Join(filepath.Dir(filePath), utils.ImgDir)
	if err := os.MkdirAll(assetsDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("create assets directory error: %v", err)
	}
	picContent := ""
	// 下载并保存图片到本地
	for i, pictureUrl := range article.Pictures {
		// 生成本地图片文件名
		ext := filepath.Ext(pictureUrl)
		if ext == "" {
			ext = ".jpg" // 默认扩展名
		}
		localFileName := fmt.Sprintf("%s_%d%s", utils.ToSafeFilename(article.Name), i, ext)
		localFilePath := filepath.Join(assetsDir, localFileName)

		log.Printf("Downloading picture in article %s: %s", article.Name, pictureUrl)
		// 使用requests下载图片
		err := requests.
			URL(pictureUrl).
			Header("user-agent", ChromeUserAgent).
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
	return picContent, nil
}
