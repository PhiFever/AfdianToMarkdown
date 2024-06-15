package album

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

func GetAlbums(authorName string) error {
	albumHost, _ := url.JoinPath(aifadian.Host, "a", authorName, "album")
	// 获取作者的所有作品集
	log.Println("albumHost:", albumHost)

	cookies := aifadian.ReadCookiesFromFile(utils.CookiePath)

	cookieString := aifadian.GetCookiesString(cookies)
	//log.Println("cookieString:", cookieString)

	userId := aifadian.GetAuthorId(authorName, albumHost, cookieString)
	//log.Println("userId:", userId)
	albumList := aifadian.GetAlbumListByInterface(userId, albumHost, cookieString)
	//log.Println("albumList:", utils.ToJSON(albumList))

	authToken := aifadian.GetAuthTokenCookieString(cookies)
	converter := md.NewConverter("", true, nil)
	for _, album := range albumList {
		//获取作品集的所有文章
		albumArticleList := aifadian.GetAlbumArticleListByInterface(album.AlbumUrl[25:], authToken)
		time.Sleep(time.Millisecond * time.Duration(aifadian.DelayMs))

		//log.Println("albumArticleList:", utils.ToJSON(albumArticleList))
		//log.Println(len(albumArticleList))

		if err := os.MkdirAll(path.Join(authorName, album.AlbumName), os.ModePerm); err != nil {
			return err
		}

		for i, article := range albumArticleList {
			filePath := path.Join(authorName, album.AlbumName, cast.ToString(i)+"_"+article.ArticleName+".md")
			log.Println("Saving file:", filePath)

			if err := aifadian.SaveContentIfNotExist(article.ArticleName, filePath, article.ArticleUrl, authToken, converter); err != nil {
				return err
			}
			//break
		}

	}
	return nil
}
