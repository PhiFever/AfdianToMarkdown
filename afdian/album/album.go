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

func GetAlbums(cfg *config.Config, authorUrlSlug string, cookieString string, authToken string, disableComment bool, quickUpdate bool) error {
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
		err := GetAlbum(cfg, cookieString, authToken, album, disableComment, quickUpdate, converter)
		if err != nil {
			return err
		}

	}
	return nil
}

func GetAlbum(cfg *config.Config, cookieString string, authToken string, album afdian.Album, disableComment bool, quickUpdate bool, converter *md.Converter) error {
	//获取作品集的所有文章
	//album.AlbumUrl会类似于 https://afdian.com/album/xyz
	re := regexp.MustCompile("^.*/album/")
	albumId := re.ReplaceAllString(album.AlbumUrl, "")

	albumInfo, err := afdian.GetAlbumInfo(cfg, albumId, cookieString)
	if err != nil {
		return err
	}
	time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))

	albumSaveDir := path.Join(cfg.DataDir, albumInfo.AuthorUrlSlug, albumInfo.AlbumName)
	if err := os.MkdirAll(albumSaveDir, os.ModePerm); err != nil {
		return fmt.Errorf("create album dir <%s> error: %v", albumSaveDir, err)
	}

	//边获取边下载
	var i int64
	for i = 0; i < albumInfo.PostCount; i += 10 {
		postList, err := afdian.GetAlbumPostPage(cfg, albumId, cookieString, i, "desc")
		if err != nil {
			return err
		}

		for _, post := range postList {
			timePrefix := post.PublishTime.Format("2006-01-02_15_04_05")
			filePath := path.Join(albumSaveDir, timePrefix+"_"+post.Name+".md")

			skipped, err := storage.SavePostIfNotExist(cfg, filePath, post, authToken, disableComment, converter)
			if err != nil {
				return err
			}
			if quickUpdate && skipped {
				slog.Info("Quick update: 检测到已存在文件，跳过剩余作品集文章", "album", albumInfo.AlbumName)
				return nil
			}
		}
		time.Sleep(time.Millisecond * time.Duration(afdian.DelayMs))
	}
	return nil
}
