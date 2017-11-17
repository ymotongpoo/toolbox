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

	"fmt"
	"github.com/tebeka/selenium"
)

const (
	TDMBProgramList = "https://tv.yahoo.co.jp/listings/23/"
	BSProgramList   = "https://tv.yahoo.co.jp/listings/bs1/"
)

const (
	ProgramWaitVisible = `//*[@id="tvpgm"]`
	ProgramPath        = `//*[@id="tvpgm"]/table/tbody/tr/td/table/tbody/tr/td/span/a`
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

func fetchPrograms(wd selenium.WebDriver) ([]Result, error) {
	if err := wd.Get(TDMBProgramList); err != nil {
		return nil, err
	}
	cond := func(wd selenium.WebDriver) (bool, error) {
		target, err := wd.FindElement(selenium.ByXPATH, ProgramWaitVisible)
		if err != nil {
			return false, err
		}
		return target.IsDisplayed()
	}
	err := wd.WaitWithTimeout(cond, 20*time.Second)
	if err != nil {
		return nil, err
	}
	elems, err := wd.FindElements(selenium.ByXPATH, ProgramPath)
	if err != nil {
		return nil, err
	}
	results := make([]Result, len(elems))
	for i, e := range elems {
		url, err := e.GetAttribute("href")
		if err != nil {
			fmt.Printf("error: %s", err)
		}
		title, err := e.Text()
		if err != nil {
			fmt.Printf("error: %s", err)
		}
		results[i] = NewResult(url, title)
	}
	return results, nil
}
