package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
)

func DownloadImages(downloadFolder string, downloadHistory string) {
	// read download history
	posts := ReadJsonFile(downloadHistory)

	for key := range posts {
		// download images
		posts := posts[key].([]interface{})
		for x := range posts {
			post := posts[x].(map[string]interface{})
			for data := range post {
				if post[data] == true {
					fmt.Println("skipping already downloaded", key)
					continue
				}
				var tmp string
				if x != 0 {
					tmp = key + "_" + strconv.Itoa(x+1)
				} else {
					tmp = key
				}
				download(tmp, data, downloadFolder)
				post[data] = true
			}
		}
	}
	WriteToJsonFile(downloadHistory, posts) // write download history
}

func download(filename string, url string, downloadFolder string) {
	fmt.Println("downloading", filename)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error downloading image:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
			return
		}
	}(resp.Body)

	out, err := os.Create(downloadFolder + "/" + filename + ".png")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer func(out *os.File) {
		err := out.Close()
		if err != nil {
			fmt.Println("Error closing file:", err)
			return
		}
	}(out)

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		fmt.Println("Error saving image:", err)
		return
	}
}
