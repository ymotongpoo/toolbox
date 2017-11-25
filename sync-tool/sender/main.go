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
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rjeczalik/notify"
	"github.com/ymotongpoo/toolbox/sync-tool"
)

var (
	fs            *flag.FlagSet
	pergeInterval *time.Duration
	secretsPath   *string
)

const (
	DefaultPergeInterval = 1 * time.Hour
)

func init() {
	fs = flag.NewFlagSet("base", flag.ExitOnError)
	pergeInterval = fs.Duration("perge", DefaultPergeInterval, "perge interval duration")
	secretsPath = fs.String("secrets", synctool.DefaultSecretsFile, "path to client_secret.json")
}

func main() {
	fs.Parse(os.Args[1:])
	c := make(chan notify.EventInfo, 1)
	eventList := []notify.Event{
		notify.InCreate,
		notify.InCloseWrite,
	}
	if err := notify.Watch(".", c, eventList...); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	tick := time.NewTicker(*pergeInterval)
	m := synctool.NewManager(synctool.DefaultSecretsFile) // TODO: replace file name with cli args
	err := m.Init()
	if err != nil {
		log.Fatalln(err)
	}

	uploadAll(m)
	for {
		select {
		case ei := <-c:
			switch {
			case ei.Event() == notify.InCloseWrite:
				log.Printf("Writing to %s is done!", ei.Path())
				go upload(m, ei.Path())
			case ei.Event() == notify.InCreate:
				log.Printf("File %s is created!", ei.Path())
			}
		case <-tick.C:
			update(m)
		}
	}
}

func uploadAll(m *synctool.Manager) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalln(err)
	}
	files, err := ioutil.ReadDir(cwd)
	for _, f := range files {
		path := filepath.Join(cwd, f.Name())
		err := upload(m, path)
		if err != nil {
			log.Println(err)
		}
	}
}

func upload(m *synctool.Manager, path string) error {
	if !strings.HasSuffix(path, ".ts") {
		return fmt.Errorf("ignoring %v from upload target.", path)
	}
	res, err := m.Upload(path, "", []string{synctool.UploadTargetFolderID})
	if err != nil {
		return err
	} else {
		log.Println(synctool.Loginfo(res))
	}
	f := synctool.NewFile(path, res.Id)
	f.Uploaded = true
	m.AddFile(f)
	return nil
}

func update(m *synctool.Manager) {
	err := m.SenderPerge()
	if err != nil {
		log.Println(err)
	}
	// TODO: SenderUpload is uploading already uploaded videos.
	// m.SenderUpload()
}
