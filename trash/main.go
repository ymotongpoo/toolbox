//    Copyright 2018 Yoshi Yamaguchi
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
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
)

const (
	timeonlyPattern  = "1504"
	daytimePatten    = "01021504"
	recordTimeFormat = "20060102T1504"
)

func parseTimeOnlyPattern(pattern string) (time.Time, error) {
	now := time.Now()
	t, err := time.Parse(timeonlyPattern, pattern)
	if err != nil {
		return time.Time{}, err
	}
	hour, min, sec := t.Clock()
	return time.Date(now.Year(), now.Month(), now.Day(), hour, min, sec, 0, time.UTC), nil
}

func parseDayTimePattern(pattern string) (time.Time, error) {
	now := time.Now()
	t, err := time.Parse(daytimePatten, pattern)
	if err != nil {
		return time.Time{}, err
	}
	hour, min, sec := t.Clock()
	_, month, day := t.Date()
	return time.Date(now.Year(), month, day, hour, min, sec, 0, time.UTC), nil
}

func removeMatchedFiles(files []os.FileInfo, format string) []os.FileInfo {
	matched := []os.FileInfo{}
	for _, f := range files {
		c := strings.Compare(f.Name(), format)
		if c <= 0 {
			matched = append(matched, f)
		}
	}
	return matched
}

func main() {
	flag.Parse()
	if flag.NArg() < 2 {
		log.Fatal("specify target time pattern and directory")
	}

	pattern := flag.Arg(0)
	var t time.Time
	var err error
	switch len(pattern) {
	case 4:
		t, err = parseTimeOnlyPattern(pattern)
	case 8:
		t, err = parseDayTimePattern(pattern)
	default:
		log.Fatalf("specified pattern is not supported: %v", pattern)
	}
	if err != nil {
		log.Fatalf("error parsing the pattern: %v", err)
	}
	format := t.Format(recordTimeFormat)

	dir := flag.Arg(1)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalf("cannot read target dir: %v", err)
	}

	matched := removeMatchedFiles(files, format)
	for _, m := range matched {
		os.Remove(m.Name())
	}
}
