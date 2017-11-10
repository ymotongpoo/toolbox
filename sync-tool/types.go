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

package synctool

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2/google"

	"google.golang.org/api/drive/v3"
	"io"
)

const (
	// UploadTargetFolderID is the ID of Google Drive folder where sender put the file to.
	UploadTargetFolderID = "1QaF-81k04ieUk4RB97PU1eALP0JnXN4S"

	// EncodeDoneFolderID is the ID of Google Drive Folder where the files are sent to after encoding process.
	EncodeDoneFolderID = "1i0GSCuF10lW1sx3A_vDGbvjAKIxPS2yM"

	// DefaultSecretsFile is the filename of JSON file where OAuth2 secrets are recorded.
	// This file is available on https://console.developers.google.com/.
	DefaultSecretsFile = "client_secret.json"

	// Template URL string to open the Drive file with ID.
	GoogleDriveOpenURL = "https://drive.google.com/open?id=%s"
)

// Manager is the wrapper of Google Drive files service.
type Manager struct {
	secrets string
	service *drive.Service
	files   []*File
}

type File struct {
	// Path to the local file.
	Path string

	// Google Drive file ID
	ID string
}

// NewManager creates Manager with OAuth2 client secrets. It must call Init() method to activate actual drive.Service.
func NewManager(secrets string) *Sender {
	return &Sender{
		secrets: secrets,
		service: nil,
	}
}

// initObject is the common function to initialize *drive.Service
func initObject(secrets string, s *drive.Service) error {
}

// Init creates http.Client based on oauth2.Config and holds drive.Service
// with hte credential.
func (m *Manager) Init() error {
	ctx := context.Background()
	b, err := ioutil.ReadFile(m.secrets)
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
	m.service = service
	return nil
}

// Upload sends a file in path to directory id in Google Drive with the description.
func (m *Manager) Upload(path, desc string, parents []string) (*drive.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	filename := filepath.Base(path)
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	dst := &drive.File{
		Name:        filename,
		Description: desc,
		Parents:     parents,
		MimeType:    mimeType,
	}
	res, err := s.service.Files.Create(dst).Media(f).Do()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (m *Manager) Download(id string) (path, error) {
	ctx := context.WithCancel(context.TODO())
	f, err := m.service.Files.Get(id).Fields("id", "name").Do()
	if err != nil {
		return "", err
	}
	// TODO: progressive download
}

// NewReceiver creates Receiver with OAuth2 client secrets. It must call Init() method to activate actual drive.Service.
func NewReceiver(secrets string) *Receiver {
	return &Receiver{
		secrets: secrets,
		service: nil,
		files:   []*File{},
	}
}

// Loginfo returns as string in the required data inside *data.File.
func Loginfo(f *drive.File) string {
	if f == nil {
		return "no object is available"
	}
	id := f.Id
	return fmt.Sprintf(GoogleDriveOpenURL, id)
}
