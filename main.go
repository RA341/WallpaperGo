package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"path/filepath"
	"wallpaperGo/files"
	"wallpaperGo/reddit"
)

const (
	coreFolder      = "./wallreddit"
	configPath      = coreFolder + "/" + "config.ini"
	downloadHistory = coreFolder + "/" + "download_history.json"
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
					normalRun(path)
					return nil
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "download",
						Value: "",
						Usage: "set a download path for wallpapers",
					},
				},
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func normalRun(downloadPath string) {
	// override existing download path from the config file
	if files.PathExists(configPath) != true || downloadPath != "" {

		if downloadPath == "" {
			fmt.Println("defaulting download path to ./downloads")
			downloadPath = "./downloads" // set default download path in case no config file exists and no download path is provided
		}

		paths := files.PathStruct{
			CoreFolder:      convertToAbsPath(coreFolder),
			ConfigPath:      convertToAbsPath(configPath),
			DownloadHistory: convertToAbsPath(downloadHistory),
			Downloads:       convertToAbsPath(downloadPath),
		}

		files.CreateSupportFiles(paths)
	}

	configFile, err := files.ReadConfig(configPath)
	if err != nil {
		log.Fatalln("Failed to read config file: ", err)
	}

	subreddit := files.ReadListFromConfig(configFile.Section("Reddit").Key("subreddit_list").String()) //load subreddit list

	accessToken, username := reddit.RetrieveTokens(configFile, configPath) // get access token
	err = reddit.RetrieveSavedPosts(accessToken, username, downloadHistory, subreddit)
	if err != nil {
		log.Fatalln("Failed to retrieve saved posts: ", err)
	}

	downloads := configFile.Section("Downloads").Key("download_path").String() // get download folder

	if downloads == "" {
		//downloads= filePicker() todo add this
		downloads = "./downloads"
	}

	//download images
	files.DownloadImages(downloads, downloadHistory)
}

func convertToAbsPath(path string) string {

	tmp, err := filepath.Abs(path)

	if err != nil {
		log.Fatalln("Failed to get absolute directory: ", err)
	}

	return tmp
}

//func filePicker() string {
//	selectedFile, err := cfdutil.ShowPickFolderDialog(cfd.DialogConfig{
//		Title:  "Pick download folder",
//		Role:   "PickDownloadFolder",
//		Folder: "C:\\",
//	})
//
//	if err == cfd.ErrorCancelled {
//		log.Fatal("Please select a folder")
//	} else if err != nil {
//		log.Fatal(err)
//	}
//
//	return selectedFile
//}
