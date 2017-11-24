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
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	UNIV            = "28"
	BSEX            = "BS01_0"
	BSTBS           = "BS01_1"
	BSNTV           = "BS13_0"
	BSCX            = "BS13_1"
	BSJPN           = "BS03_1"
	NHKBS1          = "BS15_0"
	NHKBS2          = "BS15_1"
)

var ProviderMap = map[string]Provider{
	"NHK総合1・東京":  NHK,
	"NHKEテレ1東京":  ETV,
	"日テレ1":       NTV,
	"テレビ朝日":      EX,
	"TBS1":       TBS,
	"テレビ東京1":     TX,
	"フジテレビ":      CX,
	"TOKYO　MX1":  MX,
	"放送大学1":      UNIV,
	"NHKBS1":     NHKBS1,
	"NHKBSプレミアム": NHKBS2,
	"BS日テレ":      BSNTV,
	"BS朝日1":      BSEX,
	"BS-TBS":     BSTBS,
	"BSジャパン":     BSJPN,
	"BSフジ・181":   BSCX,
}

var ReplaceCharMap = map[string]string{
	"/": "／",
	" ": "_",
	"<": "＜",
	">": "＞",
	"?": "？",
	"(": "（",
	")": "）",
	"#": "＃",
	"*": "＊",
	"$": "＄",
	"&": "＆",
	"^": "＾",
	"!": "！",
	"@": "＠",
	"%": "％",
	"+": "＋",
}

const (
	FilePrefixFormat = "20060102T1504"
	AtCmdFormat      = "0601021504.05"
	pergeTargetLine  = -4 * 24 * time.Hour
)

type Manager struct {
	programs []*Page
}

func NewManager() *Manager {
	programs := []*Page{}
	return &Manager{
		programs: programs,
	}
}

// IsRegistered checks if the program with the ID is already registered.
func (m *Manager) IsRegistered(id string) bool {
	for _, p := range m.programs {
		if p.ID == id {
			return true
		}
	}
	return false
}

// Add appends Page in programs field.
func (m *Manager) Add(p *Page) {
	m.programs = append(m.programs, p)
}

// Purge deletss obsolete data stored in programs field.
func (m *Manager) Purge() {
	deadline := time.Now().Add(pergeTargetLine)
	purged := []*Page{}
	for _, p := range m.programs {
		if p.End.After(deadline) {
			purged = append(purged, p)
		}
	}
	m.programs = purged
}

// Page is a struct to hold TV program metadata and at Job ID.
type Page struct {
	ID       string
	URL      string
	Title    string
	Provider Provider
	Start    time.Time
	End      time.Time
	AtID     int
}

// parseTime parse the string t in time expression and returns corresponding value in time.Time
func parseTime(t string) (time.Time, time.Time, error) {
	const dtptn = `([0-9]{4})年([0-9]{1,2})月([0-9]{1,2})日（(日|月|火|水|木|金|土)）\s+([0-9]{1,2})時([0-9]{1,2})分～([0-9]{1,2})時([0-9]{1,2})分`
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return time.Now(), time.Now(), err
	}
	re := regexp.MustCompile(dtptn)
	m := re.FindStringSubmatch(t)
	if len(m) != 9 {
		return time.Now(), time.Now(), fmt.Errorf("number of submatch is invalid: %d", len(m))
	}
	year, err := strconv.Atoi(m[1])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	monthInt, err := strconv.Atoi(m[2])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	month := time.Month(monthInt)
	day, err := strconv.Atoi(m[3])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	shour, err := strconv.Atoi(m[5])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	smin, err := strconv.Atoi(m[6])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	var s, e time.Time
	if shour >= 24 {
		s = time.Date(year, month, day+1, shour-24, smin, 0, 0, jst)
	} else {
		s = time.Date(year, month, day, shour, smin, 0, 0, jst)
	}

	ehour, err := strconv.Atoi(m[7])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	emin, err := strconv.Atoi(m[8])
	if err != nil {
		return time.Now(), time.Now(), err
	}
	if ehour > 24 {
		e = time.Date(year, month, day+1, ehour-24, emin, 0, 0, jst)
	} else {
		e = time.Date(year, month, day, ehour, emin, 0, 0, jst)
	}
	return s, e, nil
}

// NewPage generates a detailed program metadata instance.
func NewPage(url, title string, provider Provider, timeStr string) (*Page, error) {
	s, e, err := parseTime(timeStr)
	if err != nil {
		return nil, err
	}
	id := filepath.Base(url) // expecting Yahoo! TV Guide detailed page URL.

	return &Page{
		ID:       id,
		URL:      url,
		Title:    title,
		Provider: provider,
		Start:    s,
		End:      e,
		AtID:     0,
	}, nil
}

// Duration returns the length of period between start time and end time in seconds.
func (p *Page) Duration() int {
	return int(p.End.Sub(p.Start) / time.Second)
}

func (p *Page) Dump() string {
	prefix := p.Start.Format(FilePrefixFormat)
	filename := fmt.Sprintf("%s-%s.ts", prefix, p.Title)
	duration := strconv.Itoa(p.Duration())
	startTime := p.Start.Format(AtCmdFormat)
	recpt1Str := []string{"echo", "recpt1", "--b25", "--sid", "hd", "--strip", string(p.Provider), duration, filename, "|", "at", "-t", startTime, "\n"}
	return strings.Join(recpt1Str, " ")
}

// Book issues recpt1 command with at command support for scheduling.
func (p *Page) Book() {
	prefix := p.Start.Format(FilePrefixFormat)
	filename := fmt.Sprintf("%s-%s.ts", prefix, p.Title)
	duration := strconv.Itoa(p.Duration())
	recpt1Str := []string{"recpt1", "--b25", "--sid", "hd", "--strip", string(p.Provider), duration, filename}
	recpt1Cmd := exec.Command("echo", recpt1Str...)
	startTime := p.Start.Format(AtCmdFormat)
	atCmd := exec.Command("at", "-t", startTime)

	pr, pw := io.Pipe()
	recpt1Cmd.Stdout = pw
	atCmd.Stdin = pr
	stderr, err := atCmd.StderrPipe()
	if err != nil {
		log.Fatalln(err)
	}
	err = recpt1Cmd.Start()
	if err != nil {
		log.Fatalf("%v\n%v\n", err, strings.Join(recpt1Str, " "))
	}
	err = atCmd.Start()
	if err != nil {
		log.Fatalln(err)
	}
	recpt1Cmd.Wait()
	pw.Close()
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		id := parseAtID(line)
		if id != 0 {
			p.AtID = id
		}
	}
	err = atCmd.Wait()
	if err != nil {
		log.Fatalf("%v", err)
	}
}

// parseAtID parse the stderr of at command and find at job ID from the output.
func parseAtID(s string) int {
	pattern := regexp.MustCompile(`job ([0-9]+) at (Sun|Mon|Tue|Wed|Thu|Fri|Sat) (Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d{1,2} \d{1,2}:\d{1,2}:\d{1,2} \d{4}`)
	m := pattern.FindStringSubmatch(s)
	if len(m) > 1 {
		id, err := strconv.Atoi(m[1])
		if err != nil {
			log.Fatalln(err)
		}
		return id
	}
	return 0
}
