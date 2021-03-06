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
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	YabumiAPI = "https://yabumi.cc/api/images.json"
	UserAgent = "Yabumi-linux/Go 1.9.2"
)

// tempImagePath returns a temporary image filename based on unix timestamp.
func tempImagePath() string {
	tempDir := os.Getenv("TMPDIR")
	if tempDir == "" {
		tempDir = "/tmp"
	}
	unixNano := strconv.FormatInt(time.Now().UnixNano(), 10)
	return filepath.Join(tempDir, unixNano+".png")
}

// gnomeScreenshotPath returns the path to 'gnome-screenshot' from OS environment variable PATH.
// If gnome-screenshot is not found, it raise error.
func gnomeScreenshotPath() (string, error) {
	pathEnv := os.Getenv("PATH")
	paths := strings.Split(pathEnv, ":")
	for _, v := range paths {
		absPath := filepath.Join(v, "gnome-screenshot")
		if s, _ := os.Stat(absPath); s != nil { // TODO(ymotongpoo): os.IsExist() seems not working as expected.
			return absPath, nil
		}
	}
	return "", fmt.Errorf("gnome-screenshot is not found. please install it in your enviroment.")
}

//
func takeScreenshot() (string, error) {
	gsPath, err := gnomeScreenshotPath()
	if err != nil {
		return "", err
	}
	tmpfile := tempImagePath()
	cmd := exec.Command(gsPath, "-a", "-f", tmpfile)
	err = cmd.Run()
	if err != nil {
		return "", err
	}
	return tmpfile, nil
}

// upload sends a image file to Yabumi in form/multipart format and get status infomation.
func upload(filename string) (*http.Response, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	ff, err := w.CreateFormFile("imagedata", filename)
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(ff, f); err != nil {
		return nil, err
	}
	w.Close()

	req, err := http.NewRequest("POST", YabumiAPI, &b)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	req.Header.Set("Content-Type", w.FormDataContentType())
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// openBrowser opens the url in the default browser.
func openBrowser(url string) error {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	return err
}

func main() {
	tmpfile, err := takeScreenshot()
	if err != nil {
		log.Fatalln(err)
	}
	defer os.Remove(tmpfile)

	res, err := upload(tmpfile)
	if err != nil {
		log.Fatalln(err)
	}
	defer res.Body.Close()

	err = openBrowser(res.Header.Get("X-Yabumi-Image-Edit-Url"))
	if err != nil {
		log.Fatalln(err)
	}
}
