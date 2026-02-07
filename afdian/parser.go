package afdian

import (
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/utils"
	"fmt"
	"net/url"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/spf13/cast"
	"github.com/tidwall/gjson"
	"golang.org/x/exp/slog"
)

// GetAuthorId 获取作者的ID
// refer: https://afdian.com/a/Alice
func GetAuthorId(cfg *config.Config, authorUrlSlug string, referer string, cookieString string) (string, error) {
	apiUrl := fmt.Sprintf("%s/api/user/get-profile-by-slug?url_slug=%s", cfg.HostUrl, authorUrlSlug)
	body, err := NewRequestGet(cfg.Host, apiUrl, cookieString, referer)
	if err != nil {
		return "", err
	}
	authorId := gjson.GetBytes(body, "data.user.user_id").String()
	return authorId, nil
}

// GetMotionUrlList 获取作者的文章列表
// publish_sn获取的逻辑是第一轮请求为空，然后第二轮请求输入上一轮获取到的最后一篇文章的publish_sn，以此类推，直到获取到的publish_sn为空结束
func GetMotionUrlList(cfg *config.Config, userName string, cookieString string, prevPublishSn string) (authorArticleList []Post, nextPublishSn string, err error) {
	userReferer := fmt.Sprintf("%s/a/%s", cfg.HostUrl, userName)
	userId, err := GetAuthorId(cfg, userName, userReferer, cookieString)
	if err != nil {
		return nil, "", err
	}
	apiUrl := fmt.Sprintf("%s/api/post/get-list?user_id=%s&type=new&publish_sn=%s&per_page=10&group_id=&all=1&is_public=&plan_id=&title=&name=", cfg.HostUrl, userId, prevPublishSn)
	slog.Info("Get publish_sn apiUrl:", "url", apiUrl)

	body, err := NewRequestGet(cfg.Host, apiUrl, cookieString, userReferer)
	if err != nil {
		return nil, "", err
	}

	articleListJson := gjson.GetBytes(body, "data.list")
	articleListJson.ForEach(func(key, value gjson.Result) bool {
		articleId := value.Get("post_id").String()
		articleUrl, _ := url.JoinPath(cfg.HostUrl, "post", articleId)
		var pictures []string
		for _, result := range value.Get("pics").Array() {
			pictures = append(pictures, result.String())
		}
		publishTimeStamp := cast.ToInt64(value.Get("publish_time").String())
		publishTime := time.Unix(publishTimeStamp, 0)
		authorArticleList = append(authorArticleList, Post{
			Name:        utils.ToSafeFilename(value.Get("title").String()),
			Url:         articleUrl,
			Pictures:    pictures,
			PublishTime: publishTime,
		})
		return true
	})

	nextPublishSn = gjson.GetBytes(body, fmt.Sprintf("data.list.%d.publish_sn", len(authorArticleList)-1)).String()
	slog.Info("nextPublishSn:", "sn", nextPublishSn)
	return authorArticleList, nextPublishSn, nil
}

// GetAlbumList 获取作者的作品集列表
func GetAlbumList(cfg *config.Config, userId string, referer string, cookieString string) ([]Album, error) {
	apiUrl := fmt.Sprintf("%s/api/user/get-album-list?user_id=%s", cfg.HostUrl, userId)
	body, err := NewRequestGet(cfg.Host, apiUrl, cookieString, referer)
	if err != nil {
		return nil, err
	}
	var albumList []Album
	albumListJson := gjson.GetBytes(body, "data.list")
	albumListJson.ForEach(func(key, value gjson.Result) bool {
		albumId := value.Get("album_id").String()
		albumUrl, _ := url.JoinPath(cfg.HostUrl, "album", albumId)
		albumList = append(albumList, Album{AlbumName: value.Get("title").String(), AlbumUrl: albumUrl})
		return true
	})
	return albumList, nil
}

func GetAlbumPostList(cfg *config.Config, albumId string, cookieString string) (authorUrlSlug string, albumName string, albumPostList []Post, err error) {
	postCountApiUrl := fmt.Sprintf("%s/api/user/get-album-info?album_id=%s", cfg.HostUrl, albumId)
	referer := fmt.Sprintf("%s/album/%s", cfg.HostUrl, albumId)

	postCountBodyText, err := NewRequestGet(cfg.Host, postCountApiUrl, cookieString, referer)
	if err != nil {
		return "", "", nil, err
	}
	albumName = gjson.GetBytes(postCountBodyText, "data.album.title").String()
	postCount := gjson.GetBytes(postCountBodyText, "data.album.post_count").Int()
	authorUrlSlug = gjson.GetBytes(postCountBodyText, "data.album.user.url_slug").String()

	var i int64
	for i = 0; i < postCount; i += 10 {
		apiUrl := fmt.Sprintf("%s/api/user/get-album-post?album_id=%s&lastRank=%d&rankOrder=asc&rankField=rank", cfg.HostUrl, albumId, i)
		body, err := NewRequestGet(cfg.Host, apiUrl, cookieString, referer)
		if err != nil {
			return "", "", nil, err
		}

		albumPostListJson := gjson.GetBytes(body, "data.list")
		albumPostListJson.ForEach(func(key, value gjson.Result) bool {
			postId := value.Get("post_id").String()
			postUrl, _ := url.JoinPath(cfg.HostUrl, "album", albumId, postId)

			var pictures []string
			for _, result := range value.Get("pics").Array() {
				pictures = append(pictures, result.String())
			}
			publishTimeStamp := cast.ToInt64(value.Get("publish_time").String())
			publishTime := time.Unix(publishTimeStamp, 0)
			albumPostList = append(albumPostList, Post{
				Name:        utils.ToSafeFilename(value.Get("title").String()),
				Url:         postUrl,
				Pictures:    pictures,
				PublishTime: publishTime,
			})
			return true
		})
	}

	return authorUrlSlug, albumName, albumPostList, nil
}

// GetPostContent 获取文章正文内容
func GetPostContent(cfg *config.Config, articleUrl string, authToken string, converter *md.Converter) (string, error) {
	//在album内的： https://afdian.com/api/post/get-detail?post_id={post_id}&album_id={album_id}
	//在album外的： https://afdian.com/api/post/get-detail?post_id={post_id}&album_id=
	slog.Info("articleUrl:", "url", articleUrl)
	var apiUrl string
	splitUrl := strings.Split(articleUrl, "/")
	if strings.Contains(articleUrl, "album") {
		apiUrl = fmt.Sprintf("%s/api/post/get-detail?post_id=%s&album_id=%s", cfg.HostUrl, splitUrl[len(splitUrl)-1], splitUrl[len(splitUrl)-2])
	} else {
		apiUrl = fmt.Sprintf("%s/api/post/get-detail?post_id=%s&album_id=", cfg.HostUrl, splitUrl[len(splitUrl)-1])
	}
	slog.Debug("Get article content apiUrl:", "url", apiUrl)
	body, err := NewRequestGet(cfg.Host, apiUrl, authToken, articleUrl)
	if err != nil {
		return "", err
	}
	articleContent := gjson.GetBytes(body, "data.post.content").String()

	markdown, err := converter.ConvertString(articleContent)
	if err != nil {
		return "", fmt.Errorf("error converting HTML to Markdown: %w", err)
	}

	return markdown, nil
}

// GetPostComment 获取文章评论
// TODO:根据publish_sn获取全部评论
// https://afdian.com/api/comment/get-list?post_id={post_id}&publish_sn={publish_sn}&type=old&hot=
func GetPostComment(cfg *config.Config, articleUrl string, cookieString string) (commentString string, hotCommentString string, err error) {
	//https://afdian.com/api/comment/get-list?post_id={post_id}&publish_sn=&type=old&hot=1
	splitUrl := strings.Split(articleUrl, "/")
	postId := splitUrl[len(splitUrl)-1]
	apiUrl := fmt.Sprintf("%s/api/comment/get-list?post_id=%s&publish_sn=&type=old&hot=1", cfg.HostUrl, postId)
	slog.Debug("Get article comment apiUrl:", "url", apiUrl)

	body, err := NewRequestGet(cfg.Host, apiUrl, cookieString, articleUrl)
	if err != nil {
		return "", "", err
	}
	commentJson := gjson.GetBytes(body, "data.list")
	hotCommentJson := gjson.GetBytes(body, "data.hot_list")
	if hotCommentJson.Exists() {
		hotCommentString += "## 热评\n\n" + getCommentString(hotCommentJson)
	}

	commentString += "## 评论\n\n" + getCommentString(commentJson)

	return commentString, hotCommentString, nil
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
		i++
		return true
	})
	return commentString
}
