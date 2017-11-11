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
	"log"
	"time"

	"github.com/ymotongpoo/toolbox/sync-tool"
)

const PollInterval = 30 * time.Second

func main() {
	m := synctool.NewManager(synctool.DefaultSecretsFile)
	err := m.Init()
	if err != nil {
		log.Fatalln(err)
	}

	t := time.NewTicker(PollInterval)
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
			download(m, f)
			encode(f)
			upload(f)
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

func download(m *synctool.Manager, f *synctool.File) {
	log.Printf("start: %s\n", f.ID)
	// TODO: try later
	// n, path, err := m.Download(f.ID)
	// if err != nil {
	// 	log.Printf("download failed: %s\n", f.ID)
	// 	return
	// }
	n, path := 100, "aaaa"
	log.Printf("downloaded %v bytes: %v\n", n, path)
	f.Downloaded = true
}

func encode(f *synctool.File) {
	log.Printf("encoding: %s\n", f.Path)
	f.Encoded = true
}

func upload(f *synctool.File) {
	encodedPath := f.Path + ".mp4"
	log.Printf("uploading: %s\n", encodedPath)
}
