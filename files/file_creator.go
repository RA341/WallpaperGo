package files

import (
	"gopkg.in/ini.v1"
	"log"
	"os"
)

type PathStruct struct {
	CoreFolder      string
	ConfigPath      string
	DownloadHistory string
	Downloads       string
}

func CreateSupportFiles(paths PathStruct) {
	if PathExists(paths.CoreFolder) != true {
		err := os.Mkdir(paths.CoreFolder, 0755)
		if err != nil {
			log.Fatalln("Failed to create", paths.CoreFolder)
		}
	}

	if PathExists(paths.ConfigPath) != true {
		createConfigFile(paths.ConfigPath, paths.Downloads)
	}

	if PathExists(paths.DownloadHistory) != true {
		createDownloadHistoryFile(paths.DownloadHistory)
	}

	if PathExists(paths.Downloads) != true {
		err := os.Mkdir(paths.Downloads, 0755)
		if err != nil {
			log.Fatalln("Failed to create downloads directory\n", paths.CoreFolder, "\nwith error\n", err)
		}
	}
}

func createConfigFile(configPath string, downloadPath string) {

	values := map[string][]map[string]interface{}{
		"Reddit": {
			{"refresh_token": ""},
			{"username": ""},
			{"subreddit_list": ""},
		},
		"Downloads": {
			{"download_path": downloadPath},
		},
		"Temp": {
			{"token": ""},
			{"expires": ""},
		},
	}

	iniData := ini.Empty()

	for i := range values {
		sec, err := iniData.NewSection(i)
		if err != nil {
			log.Fatalln("failed to create section", i, err)
		}

		for _, v := range values[i] {
			for key, value := range v {
				_, err := sec.NewKey(key, value.(string))
				if err != nil {
					log.Fatalln("failed to create key", key, "with value", value, "in section", i, err)
				}
			}
		}
	}

	err := iniData.SaveTo(configPath)
	if err != nil {
		log.Fatalln("failed to save config file", configPath, err)
	}
}

func createDownloadHistoryFile(downloadHistoryPath string) {
	_, err := os.Create(downloadHistoryPath)
	if err != nil {
		log.Fatalln("failed to create file", downloadHistoryPath, " with error", err)
	}
	WriteToJsonFile(downloadHistoryPath, map[string]interface{}{})
}

func PathExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
