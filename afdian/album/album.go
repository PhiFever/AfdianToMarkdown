package album

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/utils"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/spf13/cast"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"
	"time"
)

func GetAlbums(authorName string, cookieString string, authToken string) error {
	albumHost, _ := url.JoinPath(afdian.Host, "a", authorName, "album")
	// 获取作者的所有作品集
	log.Println("albumHost:", albumHost)

	userId := afdian.GetAuthorId(authorName, albumHost, cookieString)
	//log.Println("userId:", userId)
	albumList := afdian.GetAlbumList(userId, albumHost, cookieString)
	//log.Println("albumList:", utils.ToJSON(albumList))

	converter := md.NewConverter("", true, nil)
	for _, album := range albumList {
		log.Println("Find album: ", album.AlbumName)
		//获取作品集的所有文章
		//album.AlbumUrl会类似于 https://afdian.com/album/xyz
		re := regexp.MustCompile("^.*/album/")
		albumId := re.ReplaceAllString(album.AlbumUrl, "")
		albumArticleList := afdian.GetAlbumArticleList(albumId, authToken)
		time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))

		//log.Println("albumArticleList:", utils.ToJSON(albumArticleList))
		//log.Println(len(albumArticleList))

		_ = os.MkdirAll(path.Join(authorName, album.AlbumName), os.ModePerm)

		for i, article := range albumArticleList {
			filePath := path.Join(utils.GetExecutionPath(), authorName, album.AlbumName, cast.ToString(i)+"_"+article.ArticleName+".md")
			log.Println("Saving file:", filePath)

			if err := afdian.SaveContentIfNotExist(article.ArticleName, filePath, article.ArticleUrl, authToken, converter); err != nil {
				return err
			}
			//break
		}

	}
	return nil
}
