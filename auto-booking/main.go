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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

const (
	port          = 9888
	batchInterval = 3 * 24 * time.Hour
	purgeInterval = 4 * 24 * time.Hour
)

func findChromeDriver() string {
	pathStr := os.Getenv("PATH")
	paths := strings.Split(pathStr, ":")
	for _, p := range paths {
		path := filepath.Join(p, "chromedriver")
		file, err := os.Stat(path)
		if file != nil && err == nil {
			return path
		}
	}
	return ""
}

func main() {
	batchT := time.NewTicker(batchInterval)
	purgeT := time.NewTicker(purgeInterval)
	m := NewManager()
	batch(m)
	for {
		select {
		case <-batchT.C:
			batch(m)
		case <-purgeT.C:
			m.Purge()
		}
	}
}

func batch(m *Manager) {
	// setup
	selenium.SetDebug(false)
	path := findChromeDriver()
	service, err := selenium.NewChromeDriverService(path, port)
	if err != nil {
		fmt.Println("path: ", path)
		log.Fatalln(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		log.Fatalln(err)
	}
	defer wd.Quit()

	ch := make(chan Result, 10000)
	done := make(chan bool, 1)
	go func() {
		err = fetchPrograms(wd, TDMBProgramList, ch, done)
		if err != nil {
			log.Fatalln(err)
		}
		time.Sleep(3 * time.Second)
		err = fetchPrograms(wd, BSProgramList, ch, done)
		if err != nil {
			log.Fatalln(err)
		}
		close(ch)
	}()

	<-done // TDMB
	<-done // BS

	now := time.Now().Format(FilePrefixFormat)

	file, err := os.Create(now + ".log")
	if err != nil {
		log.Println(err)
	}
	for r := range ch {
		p, err := getDetailedPage(wd, r.URL)
		if err != nil {
			fmt.Println(err)
		}
		if !m.IsRegistered(p.ID) {
			p.Book()
			if file != nil {
				fmt.Fprintf(file, "Job ID: %v -> %v %v (%v ~ %v)\n", p.AtID, p.Title, p.Provider, p.Start, p.End)
			}
			m.Add(p)
		}
		time.Sleep(5 * time.Second)
	}
}
