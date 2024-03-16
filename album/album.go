package album

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

func GetAlbums(authorName string) error {
	albumHost, _ := url.JoinPath(client.Host, "a", authorName, "album")
	// 获取作者的所有作品集
	log.Println("albumHost:", albumHost)

	cookies := client.ReadCookiesFromFile(utils.CookiePath)

	cookieString := client.GetCookiesString(cookies)
	//log.Println("cookieString:", cookieString)

	userId := client.GetAuthorId(authorName, albumHost, cookieString)
	//log.Println("userId:", userId)
	albumList := client.GetAlbumListByInterface(userId, albumHost, cookieString)
	//log.Println("albumList:", utils.ToJSON(albumList))

	authToken := client.GetAuthTokenCookieString(cookies)
	converter := md.NewConverter("", true, nil)
	for _, album := range albumList {
		//获取作品集的所有文章
		albumArticleList := client.GetAlbumArticleListByInterface(album.AlbumUrl[25:], authToken)
		log.Println("albumArticleList:", utils.ToJSON(albumArticleList))
		log.Println(len(albumArticleList))

		if err := os.MkdirAll(path.Join(authorName, album.AlbumName), os.ModePerm); err != nil {
			return err
		}

		//TODO: 下载作品集的所有文章
		for i, article := range albumArticleList {
			//覆盖保存到文件
			fileName := path.Join(authorName, album.AlbumName, cast.ToString(i)+"_"+article.ArticleName+".md")
			log.Println("Saving file:", fileName)
			_, fileExists := utils.FileExists(path.Join(authorName, album.AlbumName, fileName))
			//如果文件不存在，则下载
			if !fileExists {
				articleContent := client.GetArticleContentByInterface(article.ArticleUrl, authToken, converter)
				//log.Println("articleContent:", articleContent)
				err := os.WriteFile(fileName, []byte(articleContent), os.ModePerm)
				if err != nil {
					return err
				}
			} else {
				log.Println(fileName, "已存在，跳过下载")
			}
			time.Sleep(time.Millisecond * time.Duration(client.DelayMs))
			//break
		}

	}
	return nil
}

// Deprecated: Using getAlbumListByInterface instead
func getAlbumList(pageDoc *goquery.Document) []client.Album {
	// 获取作品集列表
	var albumList []client.Album
	//#app > div.wrapper.app-view > div > section.page-content-w100 > div > section.mt32 > div
	albumListBoxSelector := `#app > div.wrapper.app-view > div > section.page-content-w100 > div > section.mt32 > div`
	pageDoc.Find(albumListBoxSelector).Each(func(i int, albumBoxList *goquery.Selection) {
		albumSelector := `a.item`
		albumBoxList.Find(albumSelector).Each(func(i int, albumBox *goquery.Selection) {
			subUrl, _ := albumBox.Attr("href")
			albumUrl, _ := url.JoinPath(client.Host, subUrl)
			albumName := albumBox.Find(".tit.gl-hover-text-purple").Text()
			//log.Println(albumName)
			//log.Println(albumUrl)
			albumList = append(albumList, client.Album{AlbumName: albumName, AlbumUrl: albumUrl})
		})
	})

	return albumList
}
