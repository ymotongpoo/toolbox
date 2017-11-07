//    Copyright 2017 Yoshi Yamaguchi
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/drive/v3"
)

type Sender struct {
	secrets string
	service *drive.Service
}

func NewSender(secrets string) *Sender {
	return &Sender{
		secrets: secrets,
		service: nil,
	}
}

func getHTTPClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	tokenCacheDir := ".credentials"
	err := os.MkdirAll(tokenCacheDir, 0600)
	if err != nil {
		return nil, err
	}
	cache := filepath.Join(tokenCacheDir, "cache.json")
	tok, err := getToken(cache, config)
	if err != nil {
		return nil, err
	}
	err = saveToken(cache, tok)
	if err != nil {
		return nil, err
	}
	return config.Client(ctx, tok), nil
}

func getToken(cache string, config *oauth2.Config) (*oauth2.Token, error) {
	f, err := os.Open(cache)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	if err == nil {
		return t, nil
	}

	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Access the following URL in the browser and type the auth code: \n%v\n", authURL)
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, err
	}
	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func saveToken(path string, token *oauth2.Token) error {
	fmt.Printf("Saving credential to %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewEncoder(f).Encode(token)
	if err != nil {
		return err
	}
	return nil
}

func (s *Sender) Init() error {
	ctx := context.Background()
	b, err := ioutil.ReadFile(s.secrets)
	if err != nil {
		return err
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return err
	}
	client, err := getHTTPClient(ctx, config)
	if err != nil {
		return err
	}
	service, err := drive.New(client)
	if err != nil {
		return err
	}
	s.service = service
	return nil
}
