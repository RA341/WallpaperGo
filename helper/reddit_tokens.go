package helper

import (
	"fmt"
	"gopkg.in/ini.v1"
	"log"
	"strconv"
	"time"
	"wallpaperGo/reddit"
)

func RetrieveTokens(configFile *ini.File, configPath string) (string, string) {
	var accessToken string
	var tokenExpirationTime int
	var err error

	tmp := configFile.Section("Temp").Key("expires").String()
	username := configFile.Section("Reddit").Key("username").String()

	tokenExpirationTime = getTokenExpiration(tmp)

	if time.Now().Unix() > int64(tokenExpirationTime) {
		var tokens reddit.Tokens

		refreshToken := configFile.Section("Reddit").Key("refresh_token").String()

		if refreshToken == "" {
			tokens, err = reddit.Login() // login to reddit and retrieve access token
			if err != nil {
				log.Fatalln("Failed to login: ", err)
			}
		} else {
			tokens = reddit.RetrieveAccessToken(refreshToken) // get access token
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

func checkUsername(username string, tokens reddit.Tokens) reddit.Tokens {
	if username == "" {
		tokens.UserName = reddit.RetrieveUserName(tokens.AccessToken)
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

func SaveTokens(config *ini.File, configPath string, token reddit.Tokens) error {
	// this function is here because putting it in files causes a circular dependency due to reddit.Tokens

	config.Section("Reddit").Key("refresh_token").SetValue(token.RefreshToken)
	config.Section("Reddit").Key("username").SetValue(token.UserName)
	config.Section("Temp").Key("token").SetValue(token.AccessToken)
	config.Section("Temp").Key("expires").SetValue(strconv.Itoa(int(token.Timeout)))

	err := config.SaveTo(configPath)
	return err
}
