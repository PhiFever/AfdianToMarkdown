package client

import (
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
)

const (
	DelayMs = 330
	Host    = `https://afdian.net`
)

type Album struct {
	AlbumName string `json:"albumName"`
	AlbumUrl  string `json:"albumUrl"`
}

type Article struct {
	ArticleName string `json:"articleName"`
	ArticleUrl  string `json:"articleUrl"`
}

func SetAfdianHeader(req *http.Request, cookieString string, referer string) {
	req.Header.Set("authority", "afdian.net")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("afd-fe-version", "20220508")
	req.Header.Set("afd-stat-id", "c78521949a7c11ee8c2452540025c377")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("cookie", cookieString)
	req.Header.Set("dnt", "1")
	req.Header.Set("locale-lang", "zh-CN")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("referer", referer)
	req.Header.Set("sec-ch-ua", `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-gpc", "1")
	req.Header.Set("user-agent", ChromeUserAgent)
}

// NewRequestGet 发送GET请求
func NewRequestGet(Url string, cookieString string, referer string) []byte {
	client := &http.Client{}
	req, err := http.NewRequest("GET", Url, nil)
	if err != nil {
		log.Fatal(err)
	}
	SetAfdianHeader(req, cookieString, referer)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return bodyText
}

// GetAuthorId 获取作者的ID
// refer: https://afdian.net/a/q9adg
func GetAuthorId(authorName string, referer string, cookieString string) string {
	apiUrl := fmt.Sprintf("%s/api/user/get-profile-by-slug?url_slug=%s", Host, authorName)
	bodyText := NewRequestGet(apiUrl, cookieString, referer)
	//fmt.Printf("%s\n", bodyText)
	authorId := gjson.Get(string(bodyText), "data.user.user_id").String()
	return authorId
}

// GetAlbumListByInterface 获取作者的作品集列表
func GetAlbumListByInterface(userId string, referer string, cookieString string) []Album {
	apiUrl := fmt.Sprintf("%s/api/user/get-album-list?user_id=%s", Host, userId)
	bodyText := NewRequestGet(apiUrl, cookieString, referer)
	//fmt.Printf("%s\n", bodyText)
	var albumList []Album
	albumListJson := gjson.Get(string(bodyText), "data.list")
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
	postCount := gjson.Get(string(postCountBodyText), "data.album.post_count").Int()
	//log.Println("postCount:", postCount)

	var albumArticleList []Article
	var i int64
	for i = 0; i < postCount; i += 10 {
		apiUrl := fmt.Sprintf("%s/api/user/get-album-post?album_id=%s&lastRank=%d&rankOrder=asc&rankField=rank", Host, albumId, i)
		bodyText := NewRequestGet(apiUrl, authTokenCookie, referer)

		albumArticleListJson := gjson.Get(string(bodyText), "data.list")
		albumArticleListJson.ForEach(func(key, value gjson.Result) bool {
			//fmt.Println(value.Get("title").String())
			//fmt.Println(value.Get("post_id").String())
			postId := value.Get("post_id").String()
			postUrl, _ := url.JoinPath(Host, "album", albumId, postId)
			albumArticleList = append(albumArticleList, Article{ArticleName: value.Get("title").String(), ArticleUrl: postUrl})
			return true
		})
	}

	return albumArticleList
}

// GetArticleContentByInterface 获取文章内容
func GetArticleContentByInterface(articleUrl string, cookieString string, converter *md.Converter) string {
	//https://afdian.net/api/post/get-detail?post_id=0c26f170a4ea11eea1de52540025c377&album_id=c2624006a35111eeaebb52540025c377
	apiUrl := fmt.Sprintf("%s/api/post/get-detail?post_id=%s&album_id=%s", Host, articleUrl[58:], articleUrl[25:57])
	//log.Println("apiUrl:", apiUrl)
	bodyText := NewRequestGet(apiUrl, cookieString, articleUrl)
	//log.Println("bodyText: ", string(bodyText))
	articleContent := gjson.Get(string(bodyText), "data.post.content").String()
	//log.Println("articleContent: ", articleContent)

	markdown, err := converter.ConvertString(articleContent)
	if err != nil {
		log.Fatal(err)
	}
	//log.Println(markdown)
	return markdown
}
