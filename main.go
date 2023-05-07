package main

import (
	"log"
	"wallpaperGo/files"
	"wallpaperGo/reddit"
)

func main() {
	// check if local configPath files exist
	// check for flags
	// download_history.json

	coreFolder := "./wallreddit"
	downloads := "./wallpapers"
	configPath := coreFolder + "/" + "config.ini"
	downloadHistory := coreFolder + "/" + "download_history.json"

	paths := files.PathStruct{
		CoreFolder:      coreFolder,
		ConfigPath:      configPath,
		DownloadHistory: downloadHistory,
		Downloads:       downloads,
	}

	var accessToken string

	// check for files
	files.CreateSupportFiles(paths)

	// load config file
	configFile, err := files.ReadConfig(configPath)
	if err != nil {
		log.Fatalln("Failed to read config file: ", err)
	}

	username := configFile.Section("Reddit").Key("username").String() // load username

	subreddit := files.ReadListFromConfig(configFile.Section("Reddit").Key("subreddit_list").String()) //load subreddit list
	accessToken = reddit.GetAccessToken(configFile, username, configPath)                              // get access token

	err = reddit.RetrieveSavedPosts(accessToken, username, downloadHistory, subreddit)
	if err != nil {
		log.Fatalln("Failed to retrieve saved posts: ", err)
	}

	// download images
	files.DownloadImages(downloads, downloadHistory)
}
