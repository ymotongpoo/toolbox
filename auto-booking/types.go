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
	"time"
)

type Result struct {
	URL   string
	Title string
}

func NewResult(url, title string) Result {
	return Result{
		URL:   url,
		Title: title,
	}
}

type Provider string

const (
	MX     Provider = "16"
	CX              = "21"
	TBS             = "22"
	TX              = "23"
	EX              = "24"
	NTV             = "25"
	ETV             = "26"
	NHK             = "27"
	BSEX            = "BS01_0"
	BSTBS           = "BS01_1"
	BSNTV           = "BS13_0"
	BSCX            = "BS13_1"
	NHKBS1          = "BS15_0"
	NHKBS2          = "BS15_1"
)

type Page struct {
	URL      string
	Title    string
	Provider Provider
	Start    time.Time
	End      time.Time
}

// parseTime parse the string t in time expression and returns corresponding value in time.Time
func parseTime(t string) time.Time {
	return time.Now() // TODO: implement here.
}

func NewPage(url, title string, provider Provider, start, end string) Page {
	s := parseTime(start)
	e := parseTime(end)
	return Page{
		URL:      url,
		Title:    title,
		Provider: provider,
		Start:    s,
		End:      e,
	}
}

func (p Page) Duration() time.Duration {
	return p.End.Sub(p.Start)
}
