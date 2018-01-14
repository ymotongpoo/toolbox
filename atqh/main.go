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
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

const MaxAtCommand = 1000

func main() {
	ch := make(chan string, MaxAtCommand)
	atqReader(ch)
	for line := range ch {
		go atReader(line)
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

func atReader(line string) {
	elements := strings.Split(line, " ")
	id := elements[0]
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
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if !strings.HasPrefix(line, "recpt1") {
			continue
		}
		// eg. "recpt1 --b25 --sid hd --strip 26 300 20180115T0730-ピタゴラスイッチ.ts"
		recppt1Command := strings.SplitN(line, " ", 8)
		filename := recppt1Command[7]

		fmt.Printf("%v %v %v (%v) %v %v\n",
			id, elements[2], elements[3], elements[1], elements[4], filename)
	}

	if err = at.Wait(); err != nil {
		log.Fatalf("[at] failed to wait: %v\n", err)
	}
}
