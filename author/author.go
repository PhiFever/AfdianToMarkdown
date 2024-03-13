package author

import (
	"AifadianCrawler/client"
	"AifadianCrawler/utils"
	"encoding/json"
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
	cachePath = "articleUrlListCache.json"
	authorDir = "author"
)

type authorArticle struct {
	ArticleName string `json:"articleName"`
	ArticleUrl  string `json:"articleUrl"`
}

// GetAuthorArticles 获取作者的所有作品
func GetAuthorArticles(authorName string) error {
	authorHost, _ := url.JoinPath(client.Host, "a", authorName)
	//创建作者文件夹
	os.MkdirAll(path.Join(authorName, authorDir), os.ModePerm)
	log.Println("authorHost:", authorHost)

	cookies := client.ReadCookiesFromFile(utils.CookiePath)
	cookiesParam := client.ConvertCookies(cookies)
	pageCtx, pageCancel := client.InitChromedpContext(false)
	defer pageCancel()

	var articleUrlList []authorArticle
	cacheInfo, cacheExists := utils.FileExists(path.Join(authorName, authorDir, cachePath))
	//获取作者作品列表
	if cacheExists && cacheInfo.ModTime().Before(time.Now().AddDate(0, 0, -1)) {
		//如果已经有了articleUrlList.json文件，则直接读取
		file, _ := os.Open(path.Join(authorName, authorDir, cachePath))
		defer file.Close()
		err := json.NewDecoder(file).Decode(&articleUrlList)
		if err != nil {
			return err
		}
	} else {
		pageDoc := client.GetHtmlDoc(client.GetScrolledRenderedPage(pageCtx, cookiesParam, authorHost))
		//fmt.Println(pageDoc)
		articleUrlList = append(articleUrlList, getAuthorArticleUrlList(pageDoc)...)
		//保存到文件
		jsonData, _ := json.MarshalIndent(articleUrlList, "", "\t")
		file, _ := os.Create(path.Join(authorName, cachePath))
		defer file.Close()
		_, err := file.Write(jsonData)
		if err != nil {
			return err
		}
	}
	//log.Println("articleUrlList:", utils.ToJSON(articleUrlList))

	converter := md.NewConverter("", true, nil)
	for i, article := range articleUrlList {
		//覆盖保存到文件
		fileName := path.Join(authorName, authorDir, cast.ToString(len(articleUrlList)-i)+"_"+article.ArticleName+".md")
		log.Println("Saving file:", fileName)
		_, fileExists := utils.FileExists(path.Join(authorName, authorDir, cachePath))
		//如果文件不存在，则下载
		if !fileExists {
			articleDoc := client.GetHtmlDoc(client.GetScrolledRenderedPage(pageCtx, cookiesParam, article.ArticleUrl))
			articleContent := getArticleContent(articleDoc, converter)
			//log.Println("articleContent:", articleContent)
			err := os.WriteFile(fileName, []byte(articleContent), os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			log.Println(fileName, "已存在，跳过下载")
		}
		//break
	}

	return nil
}

// getAuthorArticleUrlList 获取作者作品列表
func getAuthorArticleUrlList(doc *goquery.Document) []authorArticle {
	var authorArticleList []authorArticle
	doc.Find("div.vm-block-feed").Each(func(index int, box *goquery.Selection) {
		box.Find("div.feed-content.mt16.article.pointer.unlock").Each(func(index int, el *goquery.Selection) {
			subUrl := el.Find("a").AttrOr("href", "")
			articleUrl, _ := url.JoinPath(client.Host, subUrl)
			articleName := utils.ToSafeFilename(el.Find("a").Text())
			authorArticleList = append(authorArticleList, authorArticle{ArticleName: articleName, ArticleUrl: articleUrl})
		})
	})
	return authorArticleList
}

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
