package reddit

// reference https://github.com/reddit-archive/reddit/wiki/OAuth2#retrieving-the-access-token

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"gopkg.in/ini.v1"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	connType     = "tcp"
	redirectUrl  = "http://localhost:8080"
	clientId     = "4yOpeOwLI7Z3Gk-a5eeBXg"
	authUrl      = "https://www.reddit.com/api/v1/access_token"
	userAgent    = "wallpaperGo/0.1 by descendant-of-apes"
	clientSecret = ""
)

type Tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	UserName     string `json:"username"`
	Timeout      int64  `json:"expires_in"`
}

func RetrieveTokens(configFile *ini.File, configPath string) (string, string) {
	var accessToken string
	var tokenExpirationTime int
	var err error

	tmp := configFile.Section("Temp").Key("expires").String()
	username := configFile.Section("Reddit").Key("username").String()

	tokenExpirationTime = getTokenExpiration(tmp)

	if time.Now().Unix() > int64(tokenExpirationTime) {
		var tokens Tokens

		refreshToken := configFile.Section("Reddit").Key("refresh_token").String()

		if refreshToken == "" {
			tokens, err = login() // login to reddit and retrieve access token
			if err != nil {
				log.Fatalln("Failed to login: ", err)
			}
		} else {
			tokens = retrieveAccessToken(refreshToken) // get access token
		}

		tokens = checkUsername(username, tokens)

		tokens.Timeout = time.Now().Unix() + tokens.Timeout // set access token expiry time

		err = SaveTokens(configFile, configPath, tokens)
		if err != nil {
			log.Fatalln("Failed to save tokens: ", err)
		}
		accessToken = tokens.AccessToken
		username = tokens.UserName
		fmt.Println("token", tokens)
	} else {
		accessToken = configFile.Section("Temp").Key("token").String()
	}

	return accessToken, username
}

func checkUsername(username string, tokens Tokens) Tokens {
	if username == "" {
		tokens.UserName = retrieveUserName(tokens.AccessToken)
	} else {
		tokens.UserName = username
	}
	return tokens
}

func getTokenExpiration(time string) int {
	var tokenExpirationTime int
	var err error

	if time == "" {
		tokenExpirationTime = 0
	} else {
		tokenExpirationTime, err = strconv.Atoi(time)
		if err != nil {
			log.Fatalln("Failed to convert token expiration time to int: ", err)
		}
	}
	return tokenExpirationTime
}

func login() (Tokens, error) {
	// generate random state (needed for reddit api)
	rand.Seed(time.Now().UnixNano())
	state := rand.Intn(65001)

	authUrl := "https://www.reddit.com/api/v1/" +
		"authorize?client_id=" + clientId +
		"&duration=permanent" +
		"&redirect_uri=" + redirectUrl +
		"&response_type=code" +
		"&scope=identity+history" +
		"&state=" + strconv.Itoa(state)

	err := open.Run(authUrl)
	if err != nil {
		return Tokens{}, errors.New("Error opening URL:" + err.Error())
	}

	// creating the server
	fmt.Println("Starting " + connType + " server on " + redirectUrl)
	listener, err := net.Listen(connType, "localhost:8080")
	if err != nil {
		return Tokens{}, errors.New("Error listening:" + err.Error())
	}
	defer func(l net.Listener) {
		err := l.Close()
		if err != nil {
			return
		}
	}(listener)

	con, err := listener.Accept()
	if err != nil {
		return Tokens{}, errors.New("Error connecting:" + err.Error())
	}

	buffer := make([]byte, 1024)
	n, err := con.Read(buffer)
	if err != nil {
		return Tokens{}, errors.New("Error reading:" + err.Error())
	}
	data := string(buffer[:n])

	_, err = con.Write([]byte("HTTP/1.1 200 OK\r\n\r\n<html><body><h3>Authorized, feel free to close this window</h3></body></html>"))
	if err != nil {
		return Tokens{}, err
	}

	err = con.Close()
	if err != nil {
		return Tokens{}, errors.New("Error closing connection:" + err.Error())
	}

	code := extractCode(data)
	tokens := retrieveRefreshToken(code)

	return tokens, nil
}

func extractCode(input string) string {
	re := regexp.MustCompile(`=([^ ]+)`)
	match := re.FindStringSubmatch(input)
	if len(match) > 1 {
		input = match[1]
	}

	re = regexp.MustCompile(`=([^&]+)`)
	match = re.FindStringSubmatch(input)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func retrieveRefreshToken(code string) Tokens {
	// creating form data
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectUrl)

	// converting encoded url to string
	payload := strings.NewReader(data.Encode())

	// creating request
	req, _ := http.NewRequest("POST", authUrl, payload)

	// setting headers
	req.SetBasicAuth(clientId, clientSecret)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// executing the request
	resp, _ := http.DefaultClient.Do(req)

	// reading the response
	var tokens Tokens

	err := json.NewDecoder(resp.Body).Decode(&tokens)
	if err != nil {
		log.Fatalln("Error decoding JSON response:", err)
	}

	err = resp.Body.Close()
	if err != nil {
		log.Fatalln("failed to close response body", err)
	}

	return tokens
}

func retrieveAccessToken(refreshToken string) Tokens {

	// creating form data
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	// converting encoded url to string
	payload := strings.NewReader(data.Encode())

	// creating request
	req, _ := http.NewRequest("POST", authUrl, payload)

	// setting headers
	req.SetBasicAuth(clientId, clientSecret)
	req.Header.Set("User-Agent", userAgent)

	// executing the request
	resp, _ := http.DefaultClient.Do(req)

	// reading the response
	var tokens Tokens

	err := json.NewDecoder(resp.Body).Decode(&tokens)
	if err != nil {
		log.Fatalln("Error decoding JSON response:", err)
	}

	err = resp.Body.Close()
	if err != nil {
		log.Fatalln("failed to close response body", err)
	}

	return tokens
}

func retrieveUserName(token string) string {
	meUrl := "https://oauth.reddit.com/api/v1/me.json"
	data, status := requestUrl(token, meUrl)

	if status != 200 {
		log.Fatalln("failed to retrieve username with code", status, data)
	}

	if data["name"] == nil {
		log.Fatalln("failed to retrieve username EMPTY", data)
	}
	return data["name"].(string)
}

func SaveTokens(config *ini.File, configPath string, token Tokens) error {
	// this function is here because putting it in files causes a circular dependency due to Tokens

	config.Section("Reddit").Key("refresh_token").SetValue(token.RefreshToken)
	config.Section("Reddit").Key("username").SetValue(token.UserName)
	config.Section("Temp").Key("token").SetValue(token.AccessToken)
	config.Section("Temp").Key("expires").SetValue(strconv.Itoa(int(token.Timeout)))

	err := config.SaveTo(configPath)
	return err
}
