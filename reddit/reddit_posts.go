package reddit

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"wallpaperGo/files"
)

const (
	authorization = "Bearer "
)

func RetrieveSavedPosts(token string, userName string, downloadPath string, subreddits []string) error {
	if len(subreddits) == 0 {
		log.Fatalln("No subreddits specified\nPlease add subreddits to config.ini")
	}

	apiPosts := paginateResults(token, userName)

	fmt.Println("Retrieved " + strconv.Itoa(len(apiPosts)) + " saved posts from reddit")

	downloadedPosts := files.ReadJsonFile(downloadPath)
	posts := filterResults(apiPosts, downloadedPosts, subreddits)

	files.WriteToJsonFile(downloadPath, posts)
	return nil
}

func filterResults(apiPosts []interface{}, downloadedPosts map[string]interface{}, subreddits []string) map[string]interface{} {

	for i := range apiPosts {

		data := apiPosts[i].(map[string]interface{})["data"].(map[string]interface{})

		fmt.Println("data: ", data["title"])

		// filter out posts already downloaded
		var tmpDown []string
		for i := range downloadedPosts {
			tmpDown = append(tmpDown, i)
		}

		id := data["id"].(string)
		if isItemInList(id, tmpDown) == true {
			fmt.Println("skipping already added", data["title"])
			continue
		}

		// filter out posts not in the list of subreddits
		if isItemInList(data["subreddit"].(string), subreddits) != true {
			fmt.Println("skipping", data["title"], "not in list of subreddits")
			continue
		}

		// filter out text responses
		if data["selftext"] != "" {
			fmt.Println("skipping", data["title"], "is a text post")
			continue
		}

		// gallery posts
		var tmp []interface{}
		if data["is_gallery"] == true {
			if data["media_metadata"] == nil {
				fmt.Println("skipping", data["title"], "no media metadata found")
				continue
			}
			fmt.Println("adding", data["title"])
			images := data["media_metadata"].(map[string]interface{})

			for i := range images {
				imageUrl := images[i].(map[string]interface{})["s"].(map[string]interface{})["u"].(string)
				image := CreateDownloadLink(imageUrl)
				// save to list
				tmp = append(tmp, map[string]bool{image: false})
			}
			downloadedPosts[id] = tmp
		} else {
			// single image posts
			fmt.Println("adding", data["title"])
			downloadedPosts[id] = append(tmp, map[string]bool{data["url"].(string): false})
		}
	}
	return downloadedPosts
}

func paginateResults(token string, userName string) []interface{} {
	var apiPosts []interface{}
	after := ""
	for {
		savedUrl := "https://oauth.reddit.com/user/" + userName + "/saved.json?limit=100&after=" + after

		tmpResults, status := requestUrl(token, savedUrl)
		if status != 200 {
			log.Fatalln("Error retrieving saved posts:", status, "\n", tmpResults)
		}

		posts := tmpResults["data"].(map[string]interface{})["children"].([]interface{})
		fmt.Println("posts: ", len(posts))
		if len(posts) == 0 {
			break
		}

		apiPosts = append(apiPosts, posts...)
		after = posts[len(posts)-1].(map[string]interface{})["data"].(map[string]interface{})["name"].(string)
		fmt.Println("after: ", after)
		//time.Sleep(500 * time.Millisecond)
	}
	return apiPosts
}

func CreateDownloadLink(url string) string {
	// input https://preview.redd.it/oo3e09iwkmua1.jpg?width=3840&format=pjpg&auto=...
	// output https://i.redd.it/oo3e09iwkmua1.jpg

	// removes everything after ?
	tmp := strings.Split(url, "?")[0]

	// replaces preview with i for full resolution image link
	return strings.Replace(tmp, "preview", "i", -1)
}

func isItemInList(item string, list []string) bool {
	for i := range list {
		if item == list[i] {
			return true
		}
	}
	return false
}

func requestUrl(token string, meUrl string) (map[string]interface{}, int) {
	req, err := http.NewRequest("GET", meUrl, nil)
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Add("Authorization", authorization+token)
	req.Header.Add("User-Agent", userAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(resp.Body)

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Fatal(err)
	}
	return data, resp.StatusCode
}
