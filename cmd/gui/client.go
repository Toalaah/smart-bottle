package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/toalaah/smart-bottle/pkg/build"
	"github.com/toalaah/smart-bottle/pkg/build/secrets"
)

var JWT string

type Reading struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

type ReadingsResponse struct {
	Username string    `json:"username"`
	Data     []Reading `json:"data"`
}

func Login() error {
	data := url.Values{}
	data.Set("username", secrets.BackendUsername)
	data.Set("password", secrets.BackendPassword)

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/token", build.BackendAddr), strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 200 {
		return fmt.Errorf("login failed")
	}

	defer resp.Body.Close()
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return err
	}
	JWT = tokenResponse.AccessToken
	return nil
}

func GetReadings() (ReadingsResponse, error) {
	var readings ReadingsResponse
	if JWT == "" {
		return readings, fmt.Errorf("jwt is empty, perform login first to obtain auth token")
	}
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/readings/", build.BackendAddr), nil)
	if err != nil {
		return readings, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", JWT))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return readings, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&readings)
	return readings, err
}

func PostReading(r Reading) error {
	if JWT == "" {
		return fmt.Errorf("jwt is empty, perform login first to obtain auth token")
	}
	j, err := json.Marshal(r)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/readings/", build.BackendAddr), bytes.NewReader(j))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", JWT))

	resp, err := http.DefaultClient.Do(req)
	resp.Body.Close()
	return err
}
