package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	renewToken, err := interactiveAuth()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Refresh Token: ", renewToken)
}

type SpotifyTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

func exchangeCodeForRefreshToken(clientId string, clientSecret string, authCode string, redirectUri string) (string, error) {
	body := fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s",
		authCode,
		url.QueryEscape(redirectUri))

	httpRequest, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(body))
	if err != nil {
		return "", err
	}

	httpRequest.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	httpRequest.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(clientId+":"+clientSecret)))

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return "", err
	}

	responseBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return "", err
	}

	var response SpotifyTokenResponse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return "", err
	}

	if response.AccessToken == "" {
		return "", fmt.Errorf("no token returned")
	}

	return response.RefreshToken, nil
}

func interactiveAuth() (string, error) {
	var clientId, clientSecret, redirectUri, authCode string
	fmt.Println("Please input the client ID, then the client Secret, then the app's redirect URI.")
	_, err := fmt.Scanln(&clientId)
	if err != nil {
		return "", err
	}

	_, err = fmt.Scanln(&clientSecret)
	if err != nil {
		return "", err
	}

	_, err = fmt.Scanln(&redirectUri)
	if err != nil {
		return "", err
	}

	authUrl := fmt.Sprintf("https://accounts.spotify.com/authorize?response_type=code&scope=user-read-currently-playing%%20user-read-playback-state&client_id=%s&redirect_uri=%s",
		url.QueryEscape(clientId),
		redirectUri,
	)

	fmt.Println("Authentication URL: ", authUrl)
	fmt.Println("Please input the obtained authcode when being redirected: ")

	_, err = fmt.Scanln(&authCode)
	if err != nil {
		return "", err
	}

	return exchangeCodeForRefreshToken(clientId, clientSecret, authCode, redirectUri)

}
