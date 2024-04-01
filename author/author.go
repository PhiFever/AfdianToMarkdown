package author

import (
	"AifadianCrawler/client"
	"AifadianCrawler/utils"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cast"
	"log"
	"net/url"
	"os"
	"path"
	"time"
)

const (
	authorDir = "author"
)

// GetAuthorArticles 获取作者的所有作品
func GetAuthorArticles(authorName string) error {
	authorHost, _ := url.JoinPath(client.Host, "a", authorName)
	//创建作者文件夹
	os.MkdirAll(path.Join(authorName, authorDir), os.ModePerm)
	log.Println("authorHost:", authorHost)

	cookies := client.ReadCookiesFromFile(utils.CookiePath)
	cookieString := client.GetCookiesString(cookies)

	//获取作者作品列表
	prevPublishSn := ""
	var articleList []client.Article
	for {
		//获取作者作品列表
		subArticleList, publishSn := client.GetAuthorArticleUrlListByInterface(authorName, cookieString, prevPublishSn)
		articleList = append(articleList, subArticleList...)
		prevPublishSn = publishSn
		if publishSn == "" {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(client.DelayMs))
	}
	//log.Println("articleList:", utils.ToJSON(articleList))
	//log.Println("articleList length:", len(articleList))

	converter := md.NewConverter("", true, nil)
	authToken := client.GetAuthTokenCookieString(cookies)
	for i, article := range articleList {
		filePath := path.Join(authorName, authorDir, cast.ToString(i)+"_"+article.ArticleName+".md")
		log.Println("Saving file:", filePath)
		if err := client.SaveContentIfNotExist(article.ArticleName, filePath, article.ArticleUrl, authToken, converter); err != nil {
			return err
		}
		//break
	}

	return nil
}

// Deprecated: Using GetAuthorArticleUrlListByInterface instead
// getAuthorArticleUrlList 获取作者作品列表
func getAuthorArticleUrlList(doc *goquery.Document) []client.Article {
	var authorArticleList []client.Article
	doc.Find("div.vm-block-feed").Each(func(index int, box *goquery.Selection) {
		box.Find("div.feed-content.mt16.article.pointer.unlock").Each(func(index int, el *goquery.Selection) {
			subUrl := el.Find("a").AttrOr("href", "")
			articleUrl, _ := url.JoinPath(client.Host, subUrl)
			articleName := utils.ToSafeFilename(el.Find("a").Text())
			authorArticleList = append(authorArticleList, client.Article{ArticleName: articleName, ArticleUrl: articleUrl})
		})
	})
	return authorArticleList
}

// Deprecated: Using GetArticleContentByInterface instead
// getArticleContent 获取文章正文内容
func getArticleContent(doc *goquery.Document, converter *md.Converter) string {
	//获取文章内容
	var htmlContent string
	//#app > div.wrapper.app-view > div > section.page-content-w100 > div > div.content-left.max-width-640 > div > div.feed-content.mt16.post-page.unlock > article
	contentSelector := "div.feed-content.mt16.post-page.unlock > article"
	//TODO:选取默认展开的评论
	doc.Find(contentSelector).Each(func(index int, el *goquery.Selection) {
		//获取正文的html内容
		htmlContent, _ = el.Html()
	})
	markdown, err := converter.ConvertString(htmlContent)
	if err != nil {
		log.Fatal(err)
	}
	return markdown
}
