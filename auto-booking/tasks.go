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
	"context"

	"github.com/knq/chromedp"
	"github.com/knq/chromedp/cdp"
)

const (
	TDMBProgramList = "https://tv.yahoo.co.jp/listings/23/"
	BSProgramList   = "https://tv.yahoo.co.jp/listings/bs1/"
)

const (
	ProgramWaitVisible = `//*[@id="tvpgm"]`
	ProgramPath        = `//*[@id="tvpgm"]/table/tbody/tr/td/table/tbody/tr/td/span/a`
)

func fetchPrograms(c *chromedp.CDP) ([]*cdp.Node, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// find sample of chromedp.Node()
	// https://godoc.org/github.com/knq/chromedp#Nodes
	// https://github.com/knq/chromedp/blob/master/examples/logic/main.go
	programs := []*cdp.Node{}
	tasks := chromedp.Tasks{
		chromedp.Navigate(TDMBProgramList),
		chromedp.WaitVisible(ProgramWaitVisible),
		chromedp.Nodes(ProgramPath, &programs),
	}
	err := c.Run(ctx, tasks)
	if err != nil {
		return nil, err
	}
	return programs, nil
}
