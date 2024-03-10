package author

import (
	"AifadianCrawler/client"
	"AifadianCrawler/utils"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
	"os"
	"path"
)

type authorArticle struct {
	articleName string
	articleUrl  string
}

// GetAuthorArticles 获取作者的所有作品
func GetAuthorArticles(authorName string) error {
	authorHost, _ := url.JoinPath(client.Host, "a", authorName)
	log.Println("authorHost:", authorHost)

	cookies := client.ReadCookiesFromFile(utils.CookiePath)
	cookiesParam := client.ConvertCookies(cookies)
	pageCtx, pageCancel := client.InitChromedpContext(false)
	defer pageCancel()
	pageDoc := client.GetHtmlDoc(client.GetScrolledRenderedPage(pageCtx, cookiesParam, authorHost))
	//fmt.Println(pageDoc)
	articleUrlList := getAuthorArticleUrlList(pageDoc)
	log.Println("articleUrlList:", articleUrlList)

	//创建作者文件夹
	os.Mkdir(path.Join(authorName, utils.ImgDir), os.ModePerm)
	return nil
}

func getAuthorArticleUrlList(doc *goquery.Document) []authorArticle {
	var authorArticleList []authorArticle
	doc.Find("div.vm-block-feed").Each(func(index int, box *goquery.Selection) {
		box.Find("div.feed-content.mt16.article.pointer.unlock").Each(func(index int, el *goquery.Selection) {
			subUrl := el.Find("a").AttrOr("href", "")
			articleUrl, _ := url.JoinPath(client.Host, subUrl)
			articleName := el.Find("a").Text()
			authorArticleList = append(authorArticleList, authorArticle{articleName: articleName, articleUrl: articleUrl})
		})
	})
	return authorArticleList
}
