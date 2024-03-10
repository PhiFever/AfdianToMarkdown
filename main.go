package main

import (
	"AifadianCrawler/author"
	"AifadianCrawler/utils"
	"fmt"
	"github.com/fatih/color"
	"github.com/urfave/cli/v2"
	"os"
	"time"
)

var (
	authorName    string
	albumId       string
	albumListPath string
)

func main() {
	successColor := color.New(color.Bold, color.FgGreen).FprintlnFunc()
	failColor := color.New(color.Bold, color.FgRed).FprintlnFunc()
	app := &cli.App{
		Name:      "AifadianCrawler",
		Usage:     "爱发电下载器，支持按作者或按作品集爬取数据\nGithub Link:",
		UsageText: "eg:\n	./ComicCrawler -au Jay \neg:\n\t./ComicCrawler.exe -al aaassssddd \neg:\n	./ComicCrawler.exe -l gallery_list.txt",
		Version:   "0.9.0",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "author", Aliases: []string{"au"}, Destination: &authorName, Value: "", Usage: "待下载的作者名"},
			&cli.StringFlag{Name: "album", Aliases: []string{"al"}, Destination: &albumId, Value: "", Usage: "待下载的作品集id"},
			&cli.StringFlag{Name: "list", Aliases: []string{"l"}, Destination: &albumListPath, Value: "", Usage: "待下载的作品集id列表文件，每行一个id。(不能与参数-au,-al同时使用)"},
		},
		HideHelpCommand: true,
		Action: func(c *cli.Context) error {
			//记录开始时间
			startTime := time.Now()
			var albumIdList []string

			//FIXME:分支选择逻辑有很大问题
			switch {
			case authorName == "" && albumId == "" && albumListPath == "":
				return fmt.Errorf("本程序为命令行程序，请在命令行中运行参数-h以查看帮助")
			case authorName != "" && albumId != "":
				return fmt.Errorf("参数错误，请在命令行中运行参数-h以查看帮助")
			case authorName != "":
				err := author.GetAuthorArticles(authorName)
				//TODO:不应该直接panic
				if err != nil {
					panic(err)
				}
			case albumListPath != "":
				fileContent, err := utils.ReadListFile(albumListPath)
				if err != nil {
					return err
				}
				albumIdList = append(albumIdList, fileContent...)
				fallthrough
			case albumId != "":
				albumIdList = append(albumIdList, albumId)
				err := GetAlbums(albumIdList)
				//TODO:不应该直接panic
				if err != nil {
					panic(err)
				}
			default:
				return fmt.Errorf("未知参数组合")
			}

			//记录结束时间
			endTime := time.Now()
			//计算执行时间，单位为秒
			successColor(os.Stdout, "下载完毕，共耗时:", utils.GetExecutionTime(startTime, endTime))

			return nil
		},
	}
	if err := app.Run(os.Args); err != nil {
		failColor(os.Stderr, err)
	}
}

func GetAlbums(albumIdList []string) error {
	return nil
}
