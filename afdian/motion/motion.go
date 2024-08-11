package motion

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/utils"
	"log"
	"net/url"
	"os"
	"path"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/spf13/cast"
)

const (
	authorDir = "motions"
)

// GetMotions 获取作者的所有作品
func GetMotions(authorName string) error {
	authorHost, _ := url.JoinPath(afdian.Host, "a", authorName)
	//创建作者文件夹
	os.MkdirAll(path.Join(authorName, authorDir), os.ModePerm)
	log.Println("authorHost:", authorHost)

	cookies := afdian.ReadCookiesFromFile(utils.CookiePath)
	cookieString := afdian.GetCookiesString(cookies)

	//获取作者作品列表
	prevPublishSn := ""
	var articleList []afdian.Article
	for {
		//获取作者作品列表
		subArticleList, publishSn := afdian.GetAuthorMotionUrlList(authorName, cookieString, prevPublishSn)
		articleList = append(articleList, subArticleList...)
		prevPublishSn = publishSn
		if publishSn == "" {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(30))
	}
	log.Println("articleList:", utils.ToJSON(articleList))
	log.Println("articleList length:", len(articleList))

	converter := md.NewConverter("", true, nil)
	authToken := afdian.GetAuthTokenCookieString(cookies)
	for i, article := range articleList {
		filePath := path.Join(authorName, authorDir, cast.ToString(i)+"_"+article.ArticleName+".md")
		log.Println("Saving file:", filePath)
		if err := afdian.SaveContentIfNotExist(article.ArticleName, filePath, article.ArticleUrl, authToken, converter); err != nil {
			return err
		}
		//break
	}

	return nil
}
