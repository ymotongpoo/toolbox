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
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

const MaxAtCommand = 1000

type booking struct {
	id       string
	datetime time.Time
	filename string
}

type byDatetime []booking

func (bd byDatetime) Len() int           { return len(bd) }
func (bd byDatetime) Swap(i, j int)      { bd[i], bd[j] = bd[j], bd[i] }
func (bd byDatetime) Less(i, j int) bool { return bd[i].datetime.Before(bd[j].datetime) }

type byID []booking

func (bi byID) Len() int           { return len(bi) }
func (bi byID) Swap(i, j int)      { bi[i], bi[j] = bi[j], bi[i] }
func (bi byID) Less(i, j int) bool { return bi[i].id < bi[j].id }

var option = flag.Bool("id", false, "sort by id")

func main() {
	flag.Parse()
	ch := make(chan string, MaxAtCommand)
	atqReader(ch)
	bookingCh := make(chan booking, MaxAtCommand)
	go func() {
		defer close(bookingCh)
		for line := range ch {
			atReader(line, bookingCh)
		}
	}()

	bookingList := []booking{}
	for b := range bookingCh {
		bookingList = append(bookingList, b)
	}
	if !*option {
		sort.Sort(byDatetime(bookingList))
	} else {
		sort.Sort(byID(bookingList))
	}
	if flag.NArg() == 0 {
		for _, b := range bookingList {
			fmt.Printf("%v %v %v\n", b.id, b.datetime.Format(time.ANSIC), b.filename)
		}
	} else {
		day, err := strconv.Atoi(flag.Arg(0))
		if err != nil {
			log.Fatalf("[main] failed to convert string: %v", err)
		}
		if day < 0 || day > 31 {
			log.Fatalf("[main] day should be 0-31: %v", err)
		}
		for _, b := range bookingList {
			if b.datetime.Day() == day {
				fmt.Printf("%v %v %v\n", b.id, b.datetime.Format(time.ANSIC), b.filename)
			}
		}
	}
}

// Read lines from `atq` command and pass the lines into channel.
func atqReader(ch chan<- string) {
	defer close(ch)
	atq := exec.Command("atq")
	stdout, err := atq.StdoutPipe()
	if err != nil {
		log.Fatalf("[atq] failed with error: %v\n", err)
	}
	if err = atq.Start(); err != nil {
		log.Fatalf("[atq] failed to start: %v\n", err)
	}
	reader := bufio.NewReader(stdout)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("[atq] error in read lines: %v\n", err)
		}
		ch <- line
	}
	if err = atq.Wait(); err != nil {
		log.Fatalf("[atq] failed to wait: %v\n", err)
	}
}

// Read output from at command with id given in line,
// and extract recpt1 command data.
func atReader(line string, ch chan<- booking) {
	fields := strings.Fields(line)
	id := fields[0]
	at := exec.Command("at", "-c", id)
	stdout, err := at.StdoutPipe()
	if err != nil {
		log.Printf("[at] %v\n", err)
	}
	if err = at.Start(); err != nil {
		log.Fatalf("[at] failed to start: %v\n", err)
	}

	reader := bufio.NewReader(stdout)
	for {
		l, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if !strings.HasPrefix(l, "recpt1") {
			continue
		}
		// eg. "recpt1 --b25 --sid hd --strip 26 300 20180115T0730-ピタゴラスイッチ.ts"
		recpt1Command := strings.SplitN(l, " ", 8)
		if len(recpt1Command) < 8 {
			log.Printf("[at] invalid line: %s\n", l)
			continue
		}
		filename := strings.TrimSpace(recpt1Command[7])

		datetime, err := time.Parse("2006Jan2 15:04:05",
			fmt.Sprintf("%v%v%v %v", fields[5], fields[2], fields[3], fields[4]))
		if err != nil {
			log.Fatalf("[atReader] failed to parse time: %v", err)
		}
		b := booking{
			id:       id,
			datetime: datetime,
			filename: filename,
		}
		ch <- b
	}
	if err = at.Wait(); err != nil {
		log.Fatalf("[at] failed to wait: %v\n", err)
	}
}
