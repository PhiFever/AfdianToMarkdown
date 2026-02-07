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
	"os"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/urfave/cli/v3"
	"golang.org/x/exp/slog"
)

var (
	afdianHost              string
	dataDirFlag             string
	cookiePathFlag          string
	authorUrlSlug           string
	albumUrl                string
	cookieString, authToken string
	disableComment          bool
	quickUpdate             bool
	debugMode               bool

	version string
	commit  string
	date    string

	cfg *config.Config
)

func main() {
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
			&cli.StringFlag{Name: "dir", Destination: &dataDirFlag, Value: "", Usage: "数据存储目录，默认为程序所在目录下的 data 文件夹"},
			&cli.StringFlag{Name: "cookie", Destination: &cookiePathFlag, Value: "", Usage: "cookies.json 文件路径，默认为程序所在目录下的 cookies.json"},
			&cli.BoolFlag{Name: "disable_comment", Destination: &disableComment, Value: false, Usage: "为true时不下载评论"},
			&cli.BoolFlag{Name: "debug", Destination: &debugMode, Value: false, Usage: "启用调试日志"},
		},
		Before: func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
			// 根据 --debug 参数设置日志级别
			logLevel := slog.LevelInfo
			if debugMode {
				logLevel = slog.LevelDebug
			}
			slog.SetDefault(logger.SetupLogger(logLevel))
			// 解析默认目录
			appDir, err := utils.ResolveAppDir()
			if err != nil {
				return ctx, fmt.Errorf("failed to resolve app directory: %w", err)
			}

			// 数据目录：优先使用 --dir 参数，否则使用默认值
			dataDir := dataDirFlag
			if dataDir == "" {
				dataDir = utils.DefaultDataDir(appDir)
			}

			// Cookie 路径：优先使用 --cookie 参数，否则使用默认值
			cookiePath := cookiePathFlag
			if cookiePath == "" {
				cookiePath = utils.DefaultCookiePath(appDir)
			}

			cfg = config.NewConfig(afdianHost, dataDir, cookiePath)
			var err2 error
			cookieString, authToken, err2 = afdian.GetCookies(cfg.CookiePath)
			if err2 != nil {
				return ctx, err2
			}
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
					return motion.GetMotions(cfg, authorUrlSlug, cookieString, authToken, disableComment, false)
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
					return album.GetAlbum(cfg, cookieString, authToken, afdian.Album{AlbumName: "", AlbumUrl: albumUrl}, disableComment, false, converter)
				},
			},
			{
				Name:  "albums",
				Usage: "下载指定作者的所有作品集",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "author", Aliases: []string{"au"}, Destination: &authorUrlSlug, Value: "", Usage: "待下载的作者id"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					return album.GetAlbums(cfg, authorUrlSlug, cookieString, authToken, disableComment, false)
				},
			},
			{
				Name:  "update",
				Usage: "更新所有已经下载的作者的动态和作品集",
				Flags: []cli.Flag{
					&cli.BoolFlag{Name: "quick", Destination: &quickUpdate, Value: false, Usage: "快速更新：遇到已存在的文件时跳过剩余分页"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					authors, err := utils.CheckAndListAuthors(cfg.DataDir)
					if err != nil {
						return err
					}
					for _, author := range authors {
						slog.Info("Find exist author: ", "authorName", author)
						if err := motion.GetMotions(cfg, author, cookieString, authToken, disableComment, quickUpdate); err != nil {
							return err
						}
						if err := album.GetAlbums(cfg, author, cookieString, authToken, disableComment, quickUpdate); err != nil {
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
