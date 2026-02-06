package main

import (
	"AfdianToMarkdown/afdian"
	"AfdianToMarkdown/afdian/album"
	"AfdianToMarkdown/afdian/motion"
	"AfdianToMarkdown/config"
	"AfdianToMarkdown/logger"
	"AfdianToMarkdown/utils"
	"context"
	"fmt"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/urfave/cli/v3"
	"golang.org/x/exp/slog"
	"os"
	"time"
)

var (
	afdianHost              string
	authorUrlSlug           string
	albumUrl                string
	cookieString, authToken string
	disableComment          bool

	version string
	commit  string
	date    string

	cfg *config.Config
)

func main() {
	slog.SetDefault(logger.SetupLogger())
	//记录开始时间
	startTime := time.Now()
	cmd := &cli.Command{
		Name:  "AfdianToMarkdown",
		Usage: "爱发电下载器，支持按作者或按作品集爬取数据\nGithub Link: https://github.com/PhiFever/AfdianToMarkdown",
		UsageText: "eg:\n\tAfdianToMarkdown.exe motions -au Alice \n" +
			"eg:\n\tAfdianToMarkdown.exe album -u https://afdian.com/album/aaa\n" +
			"eg:\n\tAfdianToMarkdown.exe albums -au Alice \n" +
			"eg:\n\tAfdianToMarkdown.exe update",
		Version:         fmt.Sprintf("version: %s, commit: %s, build date: %s", version, commit, date),
		HideHelpCommand: true,
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "host", Destination: &afdianHost, Value: "afdian.com", Usage: "主站域名，如访问不通可自行更改"},
			&cli.BoolFlag{Name: "disable_comment", Destination: &disableComment, Value: false, Usage: "为true时不下载评论"},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// 在这里可以根据需要做全局参数的预处理
			cfg = config.NewConfig(afdianHost, utils.DefaultCookiePath())
			cookieString, authToken = afdian.GetCookies(cfg.CookiePath)
			return ctx, nil
		},
		After: func(ctx context.Context, cmd *cli.Command) error {
			// 在这里可以根据需要做全局参数的后处理
			//记录结束时间
			endTime := time.Now()
			//计算执行时间，单位为秒
			slog.Info("处理完毕", "time cost", utils.GetExecutionTime(startTime, endTime))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:  "motions",
				Usage: "下载指定作者的所有动态",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "author", Aliases: []string{"au"}, Destination: &authorUrlSlug, Value: "", Usage: "待下载的作者id"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return motion.GetMotions(cfg, authorUrlSlug, cookieString, authToken, disableComment)
				},
			},
			{
				Name:  "album",
				Usage: "下载指定的作品集",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "url", Aliases: []string{"u"}, Destination: &albumUrl, Value: "", Usage: "待下载的作品集url"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					converter := md.NewConverter("", true, nil)
					return album.GetAlbum(cfg, cookieString, authToken, afdian.Album{AlbumName: "", AlbumUrl: albumUrl}, disableComment, converter)
				},
			},
			{
				Name:  "albums",
				Usage: "下载指定作者的所有作品集",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "author", Aliases: []string{"au"}, Destination: &authorUrlSlug, Value: "", Usage: "待下载的作者id"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return album.GetAlbums(cfg, authorUrlSlug, cookieString, authToken, disableComment)
				},
			},
			{
				Name:  "update",
				Usage: "更新所有已经下载的作者的动态和作品集",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					authors, err := utils.CheckAndListAuthors()
					if err != nil {
						return err
					}
					for _, author := range authors {
						slog.Info("Find exist author: ", "authorName", author)
						if err := motion.GetMotions(cfg, author, cookieString, authToken, disableComment); err != nil {
							return err
						}
						if err := album.GetAlbums(cfg, author, cookieString, authToken, disableComment); err != nil {
							return err
						}
					}
					return nil
				},
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			//如果没有传入任何参数，则显示帮助信息
			_ = cli.ShowAppHelp(cmd)
			return nil
		},
	}
	if err := cmd.Run(context.Background(), os.Args); err != nil {
		slog.Error(err.Error())
	}
}
