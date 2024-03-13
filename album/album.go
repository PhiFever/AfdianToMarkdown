package album

import "log"

func GetAlbums(authorName string) error {
	// 获取作者的所有作品集
	log.Println("GetAlbums: ", authorName)
	return nil
}
