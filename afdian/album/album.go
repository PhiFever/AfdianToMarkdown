package album

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/utils"
	"fmt"
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
	albumHost, _ := url.JoinPath(afdian.HostUrl, "a", authorName, "album")
	log.Println("albumHost:", albumHost)
	userId := afdian.GetAuthorId(authorName, albumHost, cookieString)
	albumList := afdian.GetAlbumList(userId, albumHost, cookieString)
	converter := md.NewConverter("", true, nil)
	for _, album := range albumList {
		log.Println("Find album: ", album.AlbumName)
		err := GetAlbum(authorName, cookieString, authToken, album, converter)
		if err != nil {
			return err
		}

	}
	return nil
}

func GetAlbum(authorName string, cookieString string, authToken string, album afdian.Album, converter *md.Converter) error {
	//获取作品集的所有文章
	//album.AlbumUrl会类似于 https://afdian.com/album/xyz
	re := regexp.MustCompile("^.*/album/")
	albumId := re.ReplaceAllString(album.AlbumUrl, "")
	albumPostList := afdian.GetAlbumPostList(albumId, cookieString)
	time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))
	if err := os.MkdirAll(path.Join(authorName, album.AlbumName), os.ModePerm); err != nil {
		return fmt.Errorf("create album dir error: %v", err)
	}

	for i, post := range albumPostList {
		filePath := path.Join(utils.GetExecutionPath(), authorName, album.AlbumName, cast.ToString(i)+"_"+post.Name+".md")
		log.Println("Saving file:", filePath)

		if err := afdian.SavePostIfNotExist(filePath, post, authToken, converter); err != nil {
			return err
		}
	}
	return nil
}
