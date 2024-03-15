package album

import (
	"AifadianCrawler/client"
	"AifadianCrawler/utils"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/url"
)

func GetAlbums(authorName string) error {
	albumHost, _ := url.JoinPath(client.Host, "a", authorName, "album")
	// 获取作者的所有作品集
	log.Println("albumHost:", albumHost)

	cookies := client.ReadCookiesFromFile(utils.CookiePath)

	// Deprecated: Using getAlbumListByInterface instead
	//cookiesParam := client.ConvertCookies(cookies)
	//pageCtx, pageCancel := client.InitChromedpContext(client.ImageEnabled)
	//defer pageCancel()
	//
	//pageDoc := client.GetHtmlDoc(client.GetScrolledRenderedPage(pageCtx, cookiesParam, albumHost))
	//albumList := getAlbumList(pageDoc)
	//log.Println("albumList:", utils.ToJSON(albumList))

	cookieString := client.GetCookieString(cookies)
	//log.Println("cookieString:", cookieString)

	userId := client.GetAuthorId(authorName, albumHost, cookieString)
	//log.Println("userId:", userId)
	albumList := client.GetAlbumListByInterface(userId, albumHost, cookieString)
	log.Println("albumList:", utils.ToJSON(albumList))
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
