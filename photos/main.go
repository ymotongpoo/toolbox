// Copyright 2019 Yoshi Yamaguchi
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	photoslib "github.com/nmrshll/google-photos-api-client-go/lib-gphotos"
	photoscli "github.com/nmrshll/google-photos-api-client-go/noserver-gphotos"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	drive "google.golang.org/api/drive/v3"
	photoslibrary "google.golang.org/api/photoslibrary/v1"
)

const (
	DefaultSecretsFile = "credentials.json"
)

var (
	sourceFolderID string
	secretsFile    string
	targetAlbumID  string
	logger         *logrus.Logger
)

func init() {
	flag.StringVar(&sourceFolderID, "source", "", "Source Google Drive Folder ID")
	flag.StringVar(&secretsFile, "ds", DefaultSecretsFile, "Path to credential JSON file")
	flag.StringVar(&targetAlbumID, "target", "", "Target Google Photos album ID")
	logger = logrus.StandardLogger()
}

func main() {
	flag.Parse()
	c, err := NewClient(secretsFile)
	if err != nil {
		logger.Fatalf("error creating drive instance: %v", err)
	}
	files, err := c.FetchDriveFileList(sourceFolderID)
	if err != nil {
		logger.Fatalf("error fetching file: %v", err)
	}
	for _, f := range files {
		prodessCopy(c, f)
	}
}

func prodessCopy(c *Client, f *drive.File) error {
	logger.Infof("[download]: %s (%s)", f.Name, f.Id)
	n, name, err := c.DownloadFile(f)
	if err != nil {
		logger.Errorf("[main] failed to download %s: %v", f.Name, err)
		return err
	}
	logger.Infof("[download]: done (%d bytes)", n)
	logger.Infof("[upload]: %v", name)
	_, err = c.UploadToPhotos(name)
	if err != nil {
		logger.Errorf("[main] failed to upload %s: %v", name, err)
		return err
	}
	logger.Infof("[upload]: done upload %s", name)
	logger.Infof("[clean up]: removing file %v", name)
	err = os.Remove(name)
	if err != nil {
		logger.Errorf("[main] failed to remove %s: %v", name, err)
		return err
	}
	logger.Infof("[clean up]: done")
	return nil
}

// ---- Google services ----

type Client struct {
	secrets   string
	driveSrv  *drive.Service
	photosSrv *photoscli.Client
}

func NewClient(secrets string) (*Client, error) {
	ctx := context.Background()
	c := Client{
		secrets: secrets,
	}
	b, err := ioutil.ReadFile(secrets)
	if err != nil {
		return nil, errors.Wrap(err, "NewClient")
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		return nil, errors.Wrap(err, "NewClient")
	}
	client, err := getHTTPClient(ctx, config)
	if err != nil {
		return nil, errors.Wrap(err, "NewClient")
	}
	driveSrv, err := drive.New(client)
	if err != nil {
		return nil, errors.Wrap(err, "NewClient")
	}
	c.driveSrv = driveSrv

	photosClient, err := photoscli.NewClient(
		photoscli.AuthenticateUser(
			photoslib.NewOAuthConfig(
				photoslib.APIAppCredentials{
					ClientID:     config.ClientID,
					ClientSecret: config.ClientSecret,
				},
			),
		),
	)
	if err != nil {
		return nil, errors.Wrap(err, "NewClient: Photos")
	}
	c.photosSrv = photosClient
	return &c, nil
}

func (c *Client) FetchDriveFileList(id string) ([]*drive.File, error) {
	files := []*drive.File{}
	query := fmt.Sprintf("'%s' in parents", id)
	fl, err := c.driveSrv.Files.List().
		Fields("nextPageToken, files(id, name, kind, mimeType)").
		Q(query).
		Do()
	if err != nil {
		return nil, errors.Wrap(err, "FetchFileList: first")
	}
	files = append(files, fl.Files...)
	logger.Infof("FetchFileList: %d files", len(files))
	for fl.NextPageToken != "" {
		fln, err := c.driveSrv.Files.List().
			Fields("nextPageToken, files(id, name, kind, mimeType)").
			Q(query).
			PageToken(fl.NextPageToken).
			Do()
		if err != nil {
			return nil, errors.Wrap(err, "FetchFileList: loop")
		}
		files = append(files, fl.Files...)
		logger.Infof("FetchFileList: %d files", len(files))
		fl = fln
	}
	return files, nil
}

func (c *Client) DownloadFile(f *drive.File) (int64, string, error) {
	file, err := os.Create(f.Name)
	if err != nil {
		logger.Errorf("failed to create file: %s", f.Name)
		return 0, "", errors.Wrap(err, "DownloadFile: create file")
	}
	defer file.Close()

	res, err := c.driveSrv.Files.Get(f.Id).Download()
	if err != nil {
		return 0, "", errors.Wrap(err, "DownloadFile: Download()")
	}
	defer res.Body.Close()
	n, err := io.Copy(file, res.Body)
	if err != nil {
		return 0, "", errors.Wrap(err, "DownloadFile: copy data")
	}
	return n, f.Name, nil
}

func (c *Client) UploadToPhotos(name string) (*photoslibrary.MediaItem, error) {
	return c.photosSrv.UploadFile(name)
}

// ---- OAuth2 ----

func getHTTPClient(ctx context.Context, config *oauth2.Config) (*http.Client, error) {
	tokenCacheDir := ".credentials"
	tokenCacheFile := filepath.Join(tokenCacheDir, "cache.json")
	_, err := os.Open(tokenCacheFile)
	if err != nil {
		err := os.MkdirAll(tokenCacheDir, 0755)
		if err != nil {
			return nil, errors.Wrap(err, "getHTTPClient")
		}
		_, err = os.Create(tokenCacheFile)
		if err != nil {
			return nil, errors.Wrap(err, "getHTTPClient")
		}
	}

	tok, err := getToken(tokenCacheFile, config)
	if err != nil {
		return nil, errors.Wrap(err, "getHTTPClient")
	}
	err = saveToken(tokenCacheFile, tok)
	if err != nil {
		return nil, errors.Wrap(err, "getHTTPClient")
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
	logger.Infof("Saving credential to %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
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
