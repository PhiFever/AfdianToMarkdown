package motion

import (
	"AifadianCrawler/aifadian"
	"AifadianCrawler/utils"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/spf13/cast"
	"log"
	"net/url"
	"os"
	"path"
	"time"
)

const (
	authorDir = "motion"
)

// GetAuthorArticles 获取作者的所有作品
func GetAuthorArticles(authorName string) error {
	authorHost, _ := url.JoinPath(aifadian.Host, "a", authorName)
	//创建作者文件夹
	os.MkdirAll(path.Join(authorName, authorDir), os.ModePerm)
	log.Println("authorHost:", authorHost)

	cookies := aifadian.ReadCookiesFromFile(utils.CookiePath)
	cookieString := aifadian.GetCookiesString(cookies)

	//获取作者作品列表
	prevPublishSn := ""
	var articleList []aifadian.Article
	for {
		//获取作者作品列表
		subArticleList, publishSn := aifadian.GetAuthorArticleUrlListByInterface(authorName, cookieString, prevPublishSn)
		articleList = append(articleList, subArticleList...)
		prevPublishSn = publishSn
		if publishSn == "" {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(aifadian.DelayMs))
	}
	log.Println("articleList:", utils.ToJSON(articleList))
	log.Println("articleList length:", len(articleList))

	converter := md.NewConverter("", true, nil)
	authToken := aifadian.GetAuthTokenCookieString(cookies)
	for i, article := range articleList {
		filePath := path.Join(authorName, authorDir, cast.ToString(i)+"_"+article.ArticleName+".md")
		log.Println("Saving file:", filePath)
		if err := aifadian.SaveContentIfNotExist(article.ArticleName, filePath, article.ArticleUrl, authToken, converter); err != nil {
			return err
		}
		//break
	}

	return nil
}
