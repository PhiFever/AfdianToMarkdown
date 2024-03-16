package author

import (
	"AifadianCrawler/client"
	"AifadianCrawler/utils"
	"encoding/json"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/spf13/cast"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)

const (
	cachePath = "articleUrlListCache.json"
	authorDir = "author"
)

// GetAuthorArticles 获取作者的所有作品
func GetAuthorArticles(authorName string) error {
	authorHost, _ := url.JoinPath(client.Host, "a", authorName)
	//创建作者文件夹
	os.MkdirAll(path.Join(authorName, authorDir), os.ModePerm)
	log.Println("authorHost:", authorHost)

	cookies := client.ReadCookiesFromFile(utils.CookiePath)
	cookiesParam := client.ConvertCookies(cookies)
	pageCtx, pageCancel := client.InitChromedpContext(client.ImageEnabled)
	defer pageCancel()

	var articleUrlList []client.Article
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

// https://afdian.net/api/post/get-list?user_id=3f49234e3e8f11eb8f6152540025c377&type=old&publish_sn=&per_page=10&group_id=&all=1&is_public=&plan_id=&title=&name=
// TODO：publish_sn无法获取
func getAuthorArticleUrlListByInterface(userId string) []client.Article {
	var authorArticleList []client.Article

	reqClient := &http.Client{}
	req, err := http.NewRequest("GET", "https://afdian.net/api/post/get-list?user_id=3f49234e3e8f11eb8f6152540025c377&type=old&publish_sn=&per_page=10&group_id=&all=1&is_public=&plan_id=&title=&name=", nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("authority", "afdian.net")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("afd-fe-version", "20220508")
	req.Header.Set("afd-stat-id", "c78521949a7c11ee8c2452540025c377")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("cookie", "_ga=GA1.1.1610556398.1702557124; auth_token=a2f5931cb036de871664e0f0df9991ec_20231214203204; _ga_6STWKR7T9E=GS1.1.1710250097.4.0.1710250097.60.0.0")
	req.Header.Set("dnt", "1")
	req.Header.Set("locale-lang", "zh-CN")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("referer", "https://afdian.net/a/q9adg")
	req.Header.Set("sec-ch-ua", `"Chromium";v="122", "Not(A:Brand";v="24", "Google Chrome";v="122"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-fetch-dest", "empty")
	req.Header.Set("sec-fetch-mode", "cors")
	req.Header.Set("sec-fetch-site", "same-origin")
	req.Header.Set("sec-gpc", "1")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.0.0 Safari/537.36")
	resp, err := reqClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", bodyText)

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
