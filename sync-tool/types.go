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
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/oauth2/google"

	"google.golang.org/api/drive/v3"
)

const (
	// UploadTargetFolderID is the ID of Google Drive folder where sender put the file to.
	UploadTargetFolderID = "1QaF-81k04ieUk4RB97PU1eALP0JnXN4S"

	// EncodeDoneFolderID is the ID of Google Drive folder where the original files are sent to after encoding process.
	EncodeDoneFolderID = "1i0GSCuF10lW1sx3A_vDGbvjAKIxPS2yM"

	// MP4TargetFolderID is the ID of Google Drive folder where encoded mp4 files are sent to.
	MP4TargetFolderID = "0B_fYUdOGrPfiUnR3aWVPMElHcjg"

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

// File holds required info for encoding management.
type File struct {
	Path        string
	ID          string
	Downloaded  bool
	Encoded     bool
	EncodedPath string
	Uploaded    bool
}

func NewFile(path, id string) *File {
	return &File{
		Path:        path,
		ID:          id,
		Downloaded:  false,
		Encoded:     false,
		EncodedPath: "",
		Uploaded:    false,
	}
}

// NewManager creates Manager with OAuth2 client secrets. It must call Init() method to activate actual drive.Service.
func NewManager(secrets string) *Manager {
	return &Manager{
		secrets: secrets,
		service: nil,
	}
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

// FindFiles get files in UploadTargetfolderid.
func (m *Manager) FindFiles() ([]drive.File, error) {
	files := []drive.File{}
	// see parameter setting on https://developers.google.com/drive/v3/web/search-parameters#fn4
	query := fmt.Sprintf("'%s' in parents", UploadTargetFolderID)
	fl, err := m.service.Files.List().Q(query).Do()
	if err != nil {
		return nil, fmt.Errorf("FindFiles: %v", err)
	}
	for _, f := range fl.Files {
		files = append(files, *f)
	}
	return files, nil
}

// FindNewFiles checks new files uploaded on Google Drive and returns those.
func (m *Manager) FindNewFiles() ([]*File, error) {
	files, err := m.FindFiles()
	if err != nil {
		return nil, fmt.Errorf("FindNewFiles: %v", err)
	}
	newFiles := []*File{}
loop:
	for _, f := range files {
		for _, mf := range m.files {
			if mf == nil {
				continue
			}
			if f.Id == mf.ID {
				continue loop
			}
		}
		nf := NewFile(f.Name, f.Id)
		newFiles = append(newFiles, nf)
	}
	m.files = append(m.files, newFiles...)
	return newFiles, nil
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
	res, err := m.service.Files.Create(dst).Media(f).Do()
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (m *Manager) SenderUpload() {
	for _, f := range m.files {
		if !f.Uploaded {
			_, err := m.Upload(f.Path, "", []string{UploadTargetFolderID})
			if err == nil {
				f.Uploaded = true
			}
		}
	}
}

func (m *Manager) AddFile(f *File) {
	m.files = append(m.files, f)
}

// Download fetches and creates a file from the path to current directory.
func (m *Manager) Download(id string) (int64, string, error) {
	//ctx, cancel := context.WithCancel(context.TODO())
	f, err := m.service.Files.Get(id).Fields("id", "name").Do()
	if err != nil {
		return 0, "", err
	}
	cwd, err := os.Getwd()
	if err != nil {
		return 0, "", err
	}
	path := filepath.Join(cwd, f.Name)
	file, err := os.Create(path)
	if err != nil {
		return 0, "", err
	}
	defer file.Close()
	res, err := m.service.Files.Get(id).Download()
	if err != nil {
		return 0, "", err
	}
	defer res.Body.Close()

	n, err := io.Copy(file, res.Body)
	mf := m.GetFile(f.Id)
	if mf == nil {
		mf = &File{
			ID:          f.Id,
			Path:        path,
			Downloaded:  true,
			Encoded:     false,
			EncodedPath: "",
			Uploaded:    false,
		}
		m.files = append(m.files, mf)
	}
	mf.Downloaded = true
	return n, path, nil
}

// Move transfers a file from upload target folder to encode done folder.
func (m *Manager) Move(id string) error {
	_, err := m.service.Files.Update(id, nil).AddParents(EncodeDoneFolderID).RemoveParents(UploadTargetFolderID).Do()
	if err != nil {
		return err
	}
	return nil
}

// Encode start encoding using ffmpeg.
func (m *Manager) Encode(id string) error {
	mf := m.GetFile(id)
	args := []string{
		"-i", fmt.Sprintf("%s", mf.Path),
		"-crf", "20.0",
		"-vcodec", "libx264", "-vf", "scale=1920:1080",
		"-preset", "slow",
		"-acodec", "aac", "-strict", "experimental",
		"-ar", "48000", "-b:a", "192k",
		"-coder", "1",
		"-flags", "+loop",
		"-cmp", "chroma", "-partitions", "+parti4x4+partp8x8+partb8x8",
		"-me_method", "hex", "-subq", "6", "-me_range", "16", "-g", "60",
		"-keyint_min", "25",
		"-sc_threshold", "35",
		"-i_qfactor", "0.71",
		"-b_strategy", "1",
		"-threads", "0",
		"-f", "mp4",
		fmt.Sprintf("%s.mp4", mf.Path),
	}
	cmd := exec.Command("ffmpeg", args...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go io.Copy(os.Stdout, stdout)
	go io.Copy(os.Stderr, stderr)
	err = cmd.Run()
	if err != nil {
		return err
	}
	mf.Encoded = true
	return nil
}

// Perge removes all processed file instance from files field and delete all processed files from file system.
func (m *Manager) Perge() error {
	left := []*File{}
	for _, f := range m.files {
		if f.Downloaded && f.Encoded {
			err := os.Remove(f.Path)
			if err != nil {
				return err
			}
			err = os.Remove(f.Path + ".mp4")
			if err != nil {
				return err
			}
			continue
		}
		left = append(left, f)
	}
	m.files = left
	return nil
}

// SenderPerge checks if the file is uploaded and in encode done folder.
// If both are satisfiled, removes the original ts file.
func (m *Manager) SenderPerge() error {
	query := fmt.Sprintf("'%s' in parents", EncodeDoneFolderID)
	fl, err := m.service.Files.List().Q(query).Do()
	if err != nil {
		return err
	}

	left := []*File{}
	// TODO: find better expression here.
	for _, mf := range m.files {
		if mf.Uploaded {
			for _, f := range fl.Files {
				if mf.ID == f.Id {
					err := os.Remove(mf.Path)
					if err != nil {
						return err
					}
				}
			}
			left = append(left, mf)
		}
	}
	m.files = left
	return nil
}

// NumFiles returns number of instance in files field.
func (m *Manager) NumFiles() int {
	return len(m.files)
}

// GetFile returns File instance with id from files field.
func (m *Manager) GetFile(id string) *File {
	for _, f := range m.files {
		if f != nil && f.ID == id {
			return f
		}
	}
	return nil
}

// Loginfo returns as string in the required data inside *data.File.
func Loginfo(f *drive.File) string {
	if f == nil {
		return "no object is available"
	}
	id := f.Id
	return fmt.Sprintf(GoogleDriveOpenURL, id)
}
