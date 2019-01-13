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
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	// BaseURL is the base URL of the G-guide web page.
	BaseURL = "https://tv.so-net.ne.jp"

	// FilePrefixFormat is the standard prefix format of the file.
	FilePrefixFormat = "20060102T1504"

	// AtCmdFormat is the format to express the datetime in `at` command.
	AtCmdFormat = "0601021504.05"

	// NumWorkers is the number of the workers to fetch the detailed web page.
	NumWorkers = 5
)

var (
	// GGuideTimePattern is the date time pattern shown in G-guide web site.
	GGuideTimePattern = regexp.MustCompile(`(\d{1,2})/(\d{1,2}) \(.*\) (\d{1,2})\:(\d{1,2}) ～ (\d{1,2})\:(\d{1,2})`)

	// JST is the default *time.Location variable to make time.Date.
	JST *time.Location

	// TitleFilter is the filter to pick only relevant programs based on program title.
	TitleIncludeFilter []*regexp.Regexp
	TitleExcludeFilter []*regexp.Regexp
)

// ReplaceCharMap is the list of characters to replace to make shell safe string.
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
	"[": "【",
	"]": "】",
}

// Provider is the enum type of the TV provider.
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
	BSTX            = "BS01_2"
	NHKBS1          = "BS15_0"
	NHKBS2          = "BS03_1"
	BS11            = "BS09_0"
	BS12            = "BS09_2"
)

// ProviderMap is the map between the name strings and the enum values.
var ProviderMap = map[string]Provider{
	"ＮＨＫ総合１・東京(Ch.1)":   NHK,
	"ＮＨＫＥテレ１・東京(Ch.2)":  ETV,
	"日テレ(Ch.4)":         NTV,
	"テレビ朝日(Ch.5)":       EX,
	"ＴＢＳ(Ch.6)":         TBS,
	"テレビ東京(Ch.7)":       TX,
	"フジテレビ(Ch.8)":       CX,
	"ＴＯＫＹＯ　ＭＸ１(Ch.9)":   MX,
	"放送大学１(Ch.12)":      UNIV,
	"ＮＨＫ ＢＳ１(Ch.1)":     NHKBS1,
	"ＮＨＫ ＢＳプレミアム(Ch.3)": NHKBS2,
	"ＢＳ日テレ(Ch.4)":       BSNTV,
	"ＢＳ朝日(Ch.5)":        BSEX,
	"ＢＳ-ＴＢＳ(Ch.6)":      BSTBS,
	"ＢＳテレ東(Ch.7)":       BSTX,
	"ＢＳフジ(Ch.8)":        BSCX,
	"BS11イレブン(Ch.11)":   BS11,
	"BS12 トゥエルビ(Ch.12)": BS12,
}

var IncludePattern = []string{
	// Seasonal
	`.*まんぷく.*`,
	`.*べっぴんさん.*`,
	`.*カーネーション.*`,
	`.*いだてん.*`,
	// Weekdays
	`.*デザインあ.*`,
	`.*ピタゴラスイッチ.*`,
	`.*Eテレ0655.*`,
	`.*Eテレ2355.*`,
	`.*地球ドラマチック.*`,
	`.*BS世界のドキュメンタリー.*`,
	`.*クローズアップ現代.*`,
	`.*日経プラス10.*`,
	`.*ねほりんぱほりん.*`,
	`.*ジョジョの奇妙な冒険.*`,
	`.*ＷＢＳ.*`,
	// Weekly
	`.*旅するスペイン語.*`,
	`.*タモリ倶楽部.*`,
	`.*ブラタモリ.*`,
	`.*鉄腕.*`,
	`.*ドキュメント72時間.*`,
	`.*ルパン三世.*`,
	`.*ゴッドタン.*`,
	`.*植物男子.*`,
	`.*家事ヤロウ.*`,
	`.*プロフェッショナル　仕事の流儀.*`,
	`.*日本の話芸.*`,
	`.*所さんの目がテン.*`,
	`.*NHKスペシャル.*`,
	`.*世界ふしぎ発見.*`,
	`.*探偵\!ナイトスクープ.*`,
	`.*世界仰天ニュース.*`,
	`.*水曜どうでしょう.*`,
	`.*マツコの知らない世界.*`,
	`.*探検バクモン.*`,
	`.*日本の話芸.*`,
	`.*世界史.*`,
	`.*日本史.*`,
	`.*地理.*`,
	`.*ビジネス基礎.*`,
	`.*家庭総合.*`,
	`.*社会と情報.*`,
	`.*簿記.*`,
	`.*将棋.*`,
	`.*サラメシ.*`,
	`.*ザ・ノンフィクション.*`,
	`.*ガイアの夜明け.*`,
	`.*刑事コロンボ.*`,
	// Irregular
	`.*BS1スペシャル.*`,
	`.*新日本風土記.*`,
	`.*落語研究会.*`,
	`.*ATP.*`,
	`.*ウィンブルドン.*`,
	`.*Why！？プログラミング.*`,
	`.*カガクノミカタ.*`,
	`.*ウルトラ重機.*`,
	`.*アメトーーク.*`,
	`.*奇跡体験！アンビリバボー.*`,
	`.*ダーウィンが来た.*`,
	// Old
	`.*バカボンのパパ.*`,
	`.*ぼくらはマンガで強くなった.*`,
	`.*ポプテピピック.*`,
	`.*花子とアン.*`,
	`.*わろてんか.*`,
	`.*あさが来た.*`,
	`.*3月のライオン.*`,
	`.*超入門！落語THE　MOVIE.*`,
	`.*MR\. BEAN.*`,
	`.*西郷どん.*`,
	`.*INGRESS.*`,
	`.*獣になれない私たち.*`,
	`.*昭和元禄落語心中.*`,
	`.*幽☆遊☆白書.*`,
	`.*マッサン.*`,
	`.*ピアノの森.*`,
	`.*半分、青い。.*`,
	`.*バキ.*`,
	`.*オイコノミア.*`,
}

var ExcludePattern = []string{
	`.*プレマップ.*`,
	`.*いじめをノックアウト.*`,
	`.*梅沢富美男と東野幸治の.*`,
}

func programTitleFilter() ([]*regexp.Regexp, []*regexp.Regexp) {
	ret := []*regexp.Regexp{}
	for _, p := range IncludePattern {
		ret = append(ret, regexp.MustCompile(p))
	}
	ret2 := []*regexp.Regexp{}
	for _, p := range ExcludePattern {
		ret2 = append(ret2, regexp.MustCompile(p))
	}
	return ret, ret2
}

func init() {
	var err error
	JST, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
	TitleIncludeFilter, TitleExcludeFilter = programTitleFilter()
}

// Program is the struct to hold TV program metadata
type Program struct {
	URL         string
	Title       string
	Start       *time.Time
	End         *time.Time
	Channel     string
	Provider    Provider
	Summary     string
	Description string
}

func (p *Program) String() string {
	return fmt.Sprintf("%s %s (%s - %s) %s", p.Provider, p.URL, p.Start, p.End, p.Title)
}

// Recpt1AtCmd generates single line command string to book recpt1 command
// at the specified time based on the value in p.
func (p *Program) Recpt1AtCmd() string {
	prefix := p.Start.Format(FilePrefixFormat)
	filename := fmt.Sprintf("%s-%s.ts", prefix, p.Title)
	duration := strconv.Itoa(int(p.End.Sub(*p.Start) / time.Second))
	startTime := p.Start.Format(AtCmdFormat)
	recpt1Str := []string{"echo", "recpt1", "--b25", "--sid", "hd", "--strip",
		string(p.Provider), duration, filename, "|", "at", "-t", startTime}
	return strings.Join(recpt1Str, " ")
}

func extractStartEndTime(t string) (*time.Time, *time.Time, error) {
	found := GGuideTimePattern.FindStringSubmatch(t)
	if len(found) != 7 {
		return nil, nil, fmt.Errorf("no time found: %v", t)
	}
	mon, err := strconv.Atoi(found[1])
	if err != nil {
		return nil, nil, err
	}
	if mon < 1 || mon > 12 {
		return nil, nil, fmt.Errorf("month is out of range: %v", mon)
	}
	day, err := strconv.Atoi(found[2])
	if err != nil {
		return nil, nil, err
	}
	starth, err := strconv.Atoi(found[3])
	if err != nil {
		return nil, nil, err
	}
	startm, err := strconv.Atoi(found[4])
	if err != nil {
		return nil, nil, err
	}
	endh, err := strconv.Atoi(found[5])
	if err != nil {
		return nil, nil, err
	}
	endm, err := strconv.Atoi(found[6])
	if err != nil {
		return nil, nil, err
	}
	now := time.Now()
	year := now.Year()
	if mon < int(now.Month()) {
		year++
	}

	start := time.Date(year, time.Month(mon), day, starth, startm, 0, 0, JST)
	end := time.Date(year, time.Month(mon), day, endh, endm, 0, 0, JST)
	if start.After(end) {
		end = end.AddDate(0, 0, 1)
	}
	return &start, &end, nil
}

func escapeTitle(t string) string {
	for k, v := range ReplaceCharMap {
		t = strings.Replace(t, k, v, -1)
	}
	return t
}

func extractProgramData(d *goquery.Document) (*Program, error) {
	titleSel := d.Find("dl.basicTxt > dd").First()
	if titleSel == nil {
		return nil, fmt.Errorf("Title not found: %v", d.Url.String())
	}
	title := titleSel.Text()
	title = strings.Replace(title, "ウェブ検索", "", -1)
	title = strings.TrimSpace(title)
	title = escapeTitle(title)

	timeSel := titleSel.Next()
	timeText := strings.Replace(timeSel.Text(), "この時間帯の番組表", "", -1)
	timeText = strings.TrimSpace(timeText)
	start, end, err := extractStartEndTime(timeText)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse time: %s", err)
	}

	chanSel := timeSel.Next()
	chanText := strings.TrimSpace(chanSel.Text())

	summarySel := d.Find("p.basicTxt").First()
	summary := strings.TrimSpace(summarySel.Text())

	descSel := summarySel.Next()
	desc := descSel.Text()

	return &Program{
		URL:         d.Url.String(),
		Title:       title,
		Start:       start,
		End:         end,
		Channel:     chanText,
		Provider:    ProviderMap[chanText],
		Summary:     summary,
		Description: desc,
	}, nil
}

func fetchDetail(wg *sync.WaitGroup, urls <-chan string, ps chan<- *Program) {
	defer wg.Done()
	for {
		url, ok := <-urls
		if !ok {
			return
		}
		doc, err := goquery.NewDocument(url)
		if err != nil {
			log.Printf("Error: couldn't get this page: %v", url)
			continue
		}
		p, err := extractProgramData(doc)
		ps <- p
		time.Sleep(1 * time.Second)
	}
}

// WriteToShellScript dumps the programs data from ps to generate
// actual shell script to book those programs.
func WriteToShellScript(ps <-chan *Program) error {
	now := time.Now().Format("20060102T1504")
	file, err := os.OpenFile(now+".sh", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Fprintln(file, "#!/bin/bash")
	for p := range ps {
		fmt.Fprintln(file, p.Recpt1AtCmd())
	}
	return nil
}

func filterProgramWithTitle(title string) bool {
	for _, p := range TitleExcludeFilter {
		if p.MatchString(title) {
			return false
		}
	}
	for _, p := range TitleIncludeFilter {
		if p.MatchString(title) {
			return true
		}
	}
	return false
}

func fetchAllProgramsStage(wg *sync.WaitGroup, ch chan<- string, d *goquery.Document) {
	defer wg.Done()
	log.Println("start fetching URLs")
	d.Find("a.schedule-link").Each(func(_ int, s *goquery.Selection) {
		title := s.Text()
		if !filterProgramWithTitle(title) {
			return
		}

		href, ok := s.Attr("href")
		if !ok {
			log.Printf("Error: URL didn't found: %v", title)
			return
		}
		ch <- BaseURL + href
	})
}

func fetchDetailStage(ch <-chan string, ps chan *Program) {
	var wg sync.WaitGroup
	for i := 0; i < NumWorkers; i++ {
		wg.Add(1)
		go fetchDetail(&wg, ch, ps)
	}
	wg.Wait()
	close(ps)
}

func main() {
	charts := []string{
		BaseURL + "/chart/23.action?span=168",
		BaseURL + "/chart/bs1.action?span=168",
	}

	urlCh := make(chan string, 100)
	go func() {
		var wg sync.WaitGroup
		for _, c := range charts {
			doc, err := goquery.NewDocument(c)
			if err != nil {
				log.Printf("Error: document not found %v", err)
				return
			}

			wg.Add(1)
			go fetchAllProgramsStage(&wg, urlCh, doc)
		}
		wg.Wait()
		close(urlCh)
	}()

	programCh := make(chan *Program)
	go fetchDetailStage(urlCh, programCh)

	if err := WriteToShellScript(programCh); err != nil {
		log.Fatalln(err)
	}
}
