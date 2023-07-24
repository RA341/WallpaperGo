package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"wallpaperGo/files"
	"wallpaperGo/helper"
	"wallpaperGo/reddit"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			{
				Name:    "go",
				Aliases: []string{"g"},
				Usage:   "download wallpapers from saved reddit posts",
				Action: func(cCtx *cli.Context) error {
					path := cCtx.String("download")
					subreddits := cCtx.String("subreddit")
					normalRun(path, subreddits)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "download",
						Aliases: []string{"d"},
						Value:   "",
						Usage:   "set a download path for wallpapers",
					},
					&cli.StringFlag{
						Name:    "subreddit",
						Aliases: []string{"s"},
						Value:   "",
						Usage:   "set subreddits to look through eg: meme,memes",
					},
				},
			},
			//{
			//	Name:    "download",
			//	Aliases: []string{"g"},
			//	Usage:   "Downloads image directly from a reddit link, eg: wallpaperGo download https://www.reddit.com/r/wallpaper/9z2j5s/4k_minimalist_mountain/",
			//	Action: func(cCtx *cli.Context) error {
			//		path := cCtx.String("download")
			//		_ = files.ReadSubredditList(cCtx.String("subreddit"))
			//		normalRun(path)
			//		return nil
			//	},
			//},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func normalRun(downloadPath string, subreddits string) {

	path, _ := helper.GetDocumentsDir()
	coreFolder := path + "/wallpaperGo"
	configPath := coreFolder + "/" + "config.ini"
	downloadHistory := coreFolder + "/" + "download_history.json"
	defaultDownloads := coreFolder + "/" + "downloads"

	// override existing download path from the config file if download path is provided
	if files.PathExists(configPath) != true || downloadPath != "" {

		if downloadPath == "" {
			fmt.Println("defaulting download path to", defaultDownloads)
			downloadPath = defaultDownloads // set default download path in case no config file exists and no download path is provided
		}

		paths := files.PathStruct{
			CoreFolder:      helper.ConvertToAbsPath(coreFolder),
			ConfigPath:      helper.ConvertToAbsPath(configPath),
			DownloadHistory: helper.ConvertToAbsPath(downloadHistory),
			Downloads:       helper.ConvertToAbsPath(downloadPath),
		}

		files.CreateSupportFiles(paths)
	}

	fmt.Println("Saving to config to: ", path)
	fmt.Println("Saving to images to: ", downloadPath)

	configFile, err := files.ReadConfig(configPath)
	if err != nil {
		log.Fatalln("Failed to read config file: ", err)
	}

	subreddit := files.ReadSubredditList(configFile.Section("Reddit").Key("subreddit_list").String()) //load subreddit list
	argSubreddits := files.ReadSubredditList(subreddits)

	if subreddit == nil && argSubreddits == nil {
		log.Fatalln("No subreddits found in subreddit list")
		return
	}

	if subreddit == nil {
		subreddit = argSubreddits
	} else {
		subreddit = append(subreddit, argSubreddits...)
	}
	fmt.Println("subreddits: ", subreddit)
	convList := files.ConvertArrayToConfigList(subreddit)
	configFile.Section("Reddit").Key("subreddit_list").SetValue(convList)

	accessToken, username := helper.RetrieveTokens(configFile, configPath)
	err = reddit.RetrieveSavedPosts(accessToken, username, downloadHistory, subreddit)
	if err != nil {
		log.Fatalln("Failed to retrieve saved posts: ", err)
	}

	downloads := configFile.Section("Downloads").Key("download_path").String() // get download folder

	if downloads == "" {
		downloads = defaultDownloads
	}

	//download images
	files.DownloadImages(downloads, downloadHistory)
}
