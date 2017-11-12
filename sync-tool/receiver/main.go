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
	"fmt"
	"log"
	"time"

	"github.com/ymotongpoo/toolbox/sync-tool"
)

const (
	PollInterval  = 10 * time.Second
	PergeInterval = 1 * time.Minute
)

func main() {
	m := synctool.NewManager(synctool.DefaultSecretsFile)
	err := m.Init()
	if err != nil {
		log.Fatalln(err)
	}

	t := time.NewTicker(PollInterval)
	pt := time.NewTicker(PergeInterval)
	ch := make(chan *synctool.File, 20)
	for {
		select {
		case c := <-t.C:
			log.Println(m.NumFiles(), c)
			checkNewFile(m, ch)
			if err != nil {
				log.Println(err)
			}
		case f := <-ch:
			download(m, f, ch)
			encode(m, f)
			upload(m, f)
		case <-pt.C:
			m.Perge()
		}
	}
}

func checkNewFile(m *synctool.Manager, ch chan<- *synctool.File) {
	files, err := m.FindNewFiles()
	if err != nil {
		log.Println(err)
		return
	}
	for _, f := range files {
		ch <- f
	}
}

func download(m *synctool.Manager, f *synctool.File, ch chan<- *synctool.File) {
	log.Printf("start: %s\n", f.ID)
	n, path, err := m.Download(f.ID)
	if err != nil {
		log.Printf("download failed: %s\n%s\n", f.ID, err)
		ch <- f
		return
	}
	log.Printf("downloaded %v bytes: %v\n", n, path)
}

func encode(m *synctool.Manager, f *synctool.File) {
	log.Printf("encoding: %s\n", f.Path)
	err := m.Encode(f.ID)
	if err != nil {
		log.Printf("encode failed: %s\n%s\n", f.ID, err)
		return
	}
	log.Printf("encoded %s\n", f.Path)
}

func upload(m *synctool.Manager, f *synctool.File) {
	encodedPath := f.Path + ".mp4"
	df, err := m.Upload(encodedPath, "", []string{synctool.UploadTargetFolderID})
	if err != nil {
		log.Printf("upload failed: %v\n%v\n", f.ID, err)
	}
	log.Printf("uploaded %v\n%v\n", encodedPath, fmt.Sprintf(synctool.GoogleDriveOpenURL, df.Id))
}
