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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

const (
	port = 9888
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
	selenium.SetDebug(true)
	path := findChromeDriver()
	service, err := selenium.NewChromeDriverService(path, port)
	if err != nil {
		fmt.Println("path: ", path)
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()
	caps := selenium.Capabilities{"browserName": "chrome"}
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	results, err := fetchPrograms(wd, TDMBProgramList)
	if err != nil {
		panic(err)
	}

	ch := make(chan Result)
	tick := time.NewTicker(5 * time.Second)
	go func(results []Result, ch chan Result) {
		for _, r := range results {
			ch <- r
		}
	}(results, ch)

	for {
		select {
		case <-tick.C:
			r := <-ch
			getDetailedPage(wd, r.URL)
		}
	}
}
