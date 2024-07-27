package album

import (
	"AifadianCrawler/afdian"
	"AifadianCrawler/utils"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/spf13/cast"
	"log"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func GetAlbums(authorName string) error {
	albumHost, _ := url.JoinPath(afdian.Host, "a", authorName, "album")
	// 获取作者的所有作品集
	log.Println("albumHost:", albumHost)

	cookies := afdian.ReadCookiesFromFile(utils.CookiePath)

	cookieString := afdian.GetCookiesString(cookies)
	//log.Println("cookieString:", cookieString)

	userId := afdian.GetAuthorId(authorName, albumHost, cookieString)
	//log.Println("userId:", userId)
	albumList := afdian.GetAlbumListByInterface(userId, albumHost, cookieString)
	//log.Println("albumList:", utils.ToJSON(albumList))

	authToken := afdian.GetAuthTokenCookieString(cookies)
	converter := md.NewConverter("", true, nil)
	for _, album := range albumList {
		log.Println("Find album: ", album.AlbumName)
		//获取作品集的所有文章
		//album.AlbumUrl会类似于 https://afdian.com/album/xyz
		albumArticleList := afdian.GetAlbumArticleListByInterface(strings.Replace(album.AlbumUrl, "https://afdian.com/album/", "", -1), authToken)
		time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))

		//log.Println("albumArticleList:", utils.ToJSON(albumArticleList))
		//log.Println(len(albumArticleList))

		if err := os.MkdirAll(path.Join(authorName, album.AlbumName), os.ModePerm); err != nil {
			return err
		}

		for i, article := range albumArticleList {
			filePath := path.Join(authorName, album.AlbumName, cast.ToString(i)+"_"+article.ArticleName+".md")
			log.Println("Saving file:", filePath)

			if err := afdian.SaveContentIfNotExist(article.ArticleName, filePath, article.ArticleUrl, authToken, converter); err != nil {
				return err
			}
			//break
		}

	}
	return nil
}
