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
	for {
		var ch <-chan string
		select {
		case c := <-t.C:
			log.Println(m.NumFiles(), c)
			ch, err = checkNewFile(m)
			if err != nil {
				log.Println(err)
			}
		case id := <-ch:
			log.Printf("start: %s\n", id)
		}
	}
}

func checkNewFile(m *synctool.Manager) (<-chan string, error) {
	files, err := m.FindNewFiles()
	if err != nil {
		return nil, err
	} else {
		for _, f := range files {
			log.Println(f)
		}
	}
	ch := make(chan string, 1)
	return ch, nil
}
