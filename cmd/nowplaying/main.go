package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type SpotifyTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

type SpotifyNowPlayingRepsonse struct {
	RepeatState          string      `json:"repeat_state"`
	ShuffleState         bool        `json:"shuffle_state"`
	ProgressMs           int         `json:"progress_ms"`
	IsPlaying            bool        `json:"is_playing"`
	CurrentlyPlayingType string      `json:"currently_playing_type"`
	Item                 TrackObject `json:"item"` //can be episode but we don't care if it's an episode
}

type TrackObject struct {
	Album      Album         `json:"album"`   //songs only
	Artists    []Artist      `json:"artists"` //songs only
	DurationMs int           `json:"duration_ms"`
	Name       string        `json:"name"`
	Id         string        `json:"id"`
	Images     []ImageObject `json:"images"`
	Show       Show          `json:"show"` //episodes only
}

type Show struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Publisher string `json:"publisher"`
}

type Album struct {
	AlbumType   string        `json:"album_type"`
	TotalTracks int           `json:"total_tracks"`
	Id          string        `json:"id"`
	Images      []ImageObject `json:"images"`
	Name        string        `json:"name"`
}

type Artist struct {
	Id     string        `json:"id"`
	Images []ImageObject `json:"images"`
	Name   string        `json:"name"`
}

type ImageObject struct {
	Url    string `json:"url"`
	Height int    `json:"height"`
	Width  int    `json:"width"`
}

func refreshToken() (string, error) {
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	refreshToken := os.Getenv("REFRESH_TOKEN")

	body := fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", refreshToken)
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
		return "", fmt.Errorf("no access token returned")
	}

	return response.AccessToken, nil
}

func getCurrentlyPlaying(token string) (SpotifyNowPlayingRepsonse, error) {
	httpRequest, err := http.NewRequest("GET", "https://api.spotify.com/v1/me/player?additional_types=episode", nil)
	if err != nil {
		return SpotifyNowPlayingRepsonse{}, err
	}
	httpRequest.Header.Add("Authorization", "Bearer "+token)

	httpResponse, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return SpotifyNowPlayingRepsonse{}, err
	}

	if httpResponse.StatusCode != 200 {
		return SpotifyNowPlayingRepsonse{}, fmt.Errorf("didnt get HTTP 200 from API")
	}

	responseBytes, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return SpotifyNowPlayingRepsonse{}, err
	}

	var response SpotifyNowPlayingRepsonse
	err = json.Unmarshal(responseBytes, &response)
	if err != nil {
		return SpotifyNowPlayingRepsonse{}, err
	}

	return response, nil
}

func runForOneHour() {
	startTime := time.Now()

	fmt.Println("Getting a new token...")
	token, err := refreshToken()
	if err != nil {
		log.Fatal(err)
	}

	for startTime.Add(time.Hour).After(time.Now()) {
		time.Sleep(30 * time.Second)

		response, err := getCurrentlyPlaying(token)
		if err != nil {
			fmt.Printf("err.Error(): %v\n", err.Error())
			continue
		}

		// we want the content to actually be playing
		if !response.IsPlaying {
			continue
		}

		// handle music and podcasts differently
		var nowPlayingString string
		switch response.CurrentlyPlayingType {
		case "track":
			// don't display album name if it's a single
			if response.Item.Album.TotalTracks > 1 {
				nowPlayingString = fmt.Sprintf("Now Playing: %s - %s (in %s)",
					response.Item.Artists[0].Name,
					response.Item.Name,
					response.Item.Album.Name)
			} else {
				nowPlayingString = fmt.Sprintf("Now Playing: %s - %s",
					response.Item.Artists[0].Name,
					response.Item.Name)
			}
		case "episode":
			nowPlayingString = fmt.Sprintf("You are now listening to: 102.3, REAL %s FM, where we play NOTHING but %s",
				response.Item.Show.Name,
				response.Item.Name)
		}
		fmt.Printf("nowPlayingString: %v\n", nowPlayingString)

		httpRequest, err := http.NewRequest("POST", os.Getenv("NTFY_URL"), strings.NewReader(nowPlayingString))
		if err != nil {
			continue
		}
		http.DefaultClient.Do(httpRequest)
	}

}

func main() {
	for {
		runForOneHour()
	}
}
