package album

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/storage"
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"golang.org/x/exp/slog"
)

func GetAlbums(cfg *config.Config, authorUrlSlug string, cookieString string, authToken string, disableComment bool) error {
	albumHost, _ := url.JoinPath(cfg.HostUrl, "a", authorUrlSlug, "album")
	slog.Info("album列表页:", "albumHostUrl", albumHost)
	userId, err := afdian.GetAuthorId(cfg, authorUrlSlug, albumHost, cookieString)
	if err != nil {
		return err
	}
	albumList, err := afdian.GetAlbumList(cfg, userId, albumHost, cookieString)
	if err != nil {
		return err
	}
	converter := md.NewConverter("", true, nil)
	for _, album := range albumList {
		slog.Info("Find album: ", "albumName", album.AlbumName)
		err := GetAlbum(cfg, cookieString, authToken, album, disableComment, converter)
		if err != nil {
			return err
		}

	}
	return nil
}

func GetAlbum(cfg *config.Config, cookieString string, authToken string, album afdian.Album, disableComment bool, converter *md.Converter) error {
	//获取作品集的所有文章
	//album.AlbumUrl会类似于 https://afdian.com/album/xyz
	re := regexp.MustCompile("^.*/album/")
	albumId := re.ReplaceAllString(album.AlbumUrl, "")
	authorUrlSlug, albumName, albumPostList, err := afdian.GetAlbumPostList(cfg, albumId, cookieString)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))

	albumSaveDir := path.Join(cfg.DataDir, authorUrlSlug, albumName)
	if err := os.MkdirAll(albumSaveDir, os.ModePerm); err != nil {
		return fmt.Errorf("create album dir <%s> error: %v", albumSaveDir, err)
	}

	for _, post := range albumPostList {
		timePrefix := post.PublishTime.Format("2006-01-02_15_04_05")
		filePath := path.Join(albumSaveDir, timePrefix+"_"+post.Name+".md")

		if err := storage.SavePostIfNotExist(cfg, filePath, post, authToken, disableComment, converter); err != nil {
			return err
		}
	}
	return nil
}
