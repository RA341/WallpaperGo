package reddit

// reference https://github.com/reddit-archive/reddit/wiki/OAuth2#retrieving-the-access-token

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/skratchdot/open-golang/open"
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

func Login() (Tokens, error) {
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
	tokens := RetrieveRefreshToken(code)

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

func RetrieveRefreshToken(code string) Tokens {
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

func RetrieveAccessToken(refreshToken string) Tokens {

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

func RetrieveUserName(token string) string {
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
