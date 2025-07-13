package album

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/utils"
	"fmt"
	"golang.org/x/exp/slog"
	"net/url"
	"os"
	"path"
	"regexp"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/spf13/cast"
)

func GetAlbums(authorUrlSlug string, cookieString string, authToken string, disableComment bool) error {
	albumHost, _ := url.JoinPath(afdian.HostUrl, "a", authorUrlSlug, "album")
	slog.Info("album列表页:", "albumHostUrl", albumHost)
	userId := afdian.GetAuthorId(authorUrlSlug, albumHost, cookieString)
	albumList := afdian.GetAlbumList(userId, albumHost, cookieString)
	converter := md.NewConverter("", true, nil)
	for _, album := range albumList {
		slog.Info("Find album: ", "albumName", album.AlbumName)
		err := GetAlbum(cookieString, authToken, album, disableComment, converter)
		if err != nil {
			return err
		}

	}
	return nil
}

func GetAlbum(cookieString string, authToken string, album afdian.Album, disableComment bool, converter *md.Converter) error {
	//获取作品集的所有文章
	//album.AlbumUrl会类似于 https://afdian.com/album/xyz
	re := regexp.MustCompile("^.*/album/")
	albumId := re.ReplaceAllString(album.AlbumUrl, "")
	authorUrlSlug, albumName, albumPostList := afdian.GetAlbumPostList(albumId, cookieString)
	time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))

	albumSaveDir := path.Join(authorUrlSlug, albumName)
	if err := os.MkdirAll(albumSaveDir, os.ModePerm); err != nil {
		return fmt.Errorf("create album dir <%s> error: %v", albumSaveDir, err)
	}

	for i, post := range albumPostList {
		filePath := path.Join(utils.GetAppDataPath(), albumSaveDir, cast.ToString(i)+"_"+post.Name+".md")

		if err := afdian.SavePostIfNotExist(filePath, post, authToken, disableComment, converter); err != nil {
			return err
		}
	}
	return nil
}
