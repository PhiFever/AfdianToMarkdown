package album

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/utils"
	"log"
	"net/url"
	"os"
	"path"
	"regexp"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/spf13/cast"
)

func GetAlbums(authorName string, cookieString string, authToken string) error {
	albumHost, _ := url.JoinPath(afdian.Host, "a", authorName, "album")
	log.Println("albumHost:", albumHost)
	userId := afdian.GetAuthorId(authorName, albumHost, cookieString)
	albumList := afdian.GetAlbumList(userId, albumHost, cookieString)
	converter := md.NewConverter("", true, nil)
	for _, album := range albumList {
		log.Println("Find album: ", album.AlbumName)
		//获取作品集的所有文章
		//album.AlbumUrl会类似于 https://afdian.com/album/xyz
		re := regexp.MustCompile("^.*/album/")
		albumId := re.ReplaceAllString(album.AlbumUrl, "")
		albumPostList := afdian.GetAlbumPostList(albumId, cookieString)
		time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))
		_ = os.MkdirAll(path.Join(authorName, album.AlbumName), os.ModePerm)

		for i, post := range albumPostList {
			filePath := path.Join(utils.GetExecutionPath(), authorName, album.AlbumName, cast.ToString(i)+"_"+post.GetName()+".md")
			log.Println("Saving file:", filePath)

			if manga, ok := post.(afdian.Manga); ok {
				// 如果转换成功，说明当前元素是 Manga 类型
				//fmt.Println("Manga Name:", manga.Name)
				//fmt.Println("Manga URL:", manga.Url)
				//fmt.Println("Manga Pictures:", manga.Pictures)
				if err := afdian.SaveMangaIfNotExist(filePath, manga, authToken, converter); err != nil {
					return err
				}
			} else if article, ok := post.(afdian.Article); ok {
				if err := afdian.SaveContentIfNotExist(article.Name, filePath, article.Url, authToken, converter); err != nil {
					return err
				}
			} else {
				log.Fatal("Unknown post type")
			}
		}

	}
	return nil
}
