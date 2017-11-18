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
	"strings"
	"time"

	"github.com/tebeka/selenium"
)

const Timeout = 20 * time.Second

const (
	TDMBProgramList = "https://tv.yahoo.co.jp/listings/23/"
	BSProgramList   = "https://tv.yahoo.co.jp/listings/bs1/"
)

const (
	ProgramWaitVisible    = `//*[@id="tvpgm"]`
	ProgramPath           = `//*[@id="tvpgm"]/table/tbody/tr/td/table/tbody/tr/td/span/a`
	ProgramNextPage       = `//*[@id="nexttime"]/a`
	DetailedPageTitle     = `//*[@id="main"]/div[1]/div/div/div[1]/h2/b`
	DetailedPageTime      = `//*[@id="main"]/div[1]/div/div/div[1]/p/em`
	DetailedPageProvider  = `//*[@itemprop="provider"]`
	DetailedPageProvider2 = `//*[@id="yjContentsBody"]/div[1]/div[1]/p`
)

// WaitForXPath
func WaitForXPath(xpath string) selenium.Condition {
	cond := func(wd selenium.WebDriver) (bool, error) {
		target, err := wd.FindElement(selenium.ByXPATH, xpath)
		if err != nil {
			return false, err
		}
		return target.IsDisplayed()
	}
	return cond
}

func filterProgramWithTitle(r Result) bool {
	patterns := programTitleFilter()
	for _, p := range patterns {
		if p.MatchString(r.Title) {
			return true
		}
	}
	return false
}

// fetchProgramsOn get all the URLs to the program detail pages and iss link title on a program list.
func fetchProgramsOn(wd selenium.WebDriver, url string, ch chan<- Result) error {
	log.Println(url)
	if err := wd.Get(url); err != nil {
		return err
	}
	err := wd.WaitWithTimeout(WaitForXPath(ProgramWaitVisible), Timeout)
	if err != nil {
		return err
	}
	elems, err := wd.FindElements(selenium.ByXPATH, ProgramPath)
	if err != nil {
		return err
	}
	for _, e := range elems {
		url, err := e.GetAttribute("href")
		if err != nil {
			fmt.Printf("href error: %s\n", err)
		}
		title, err := e.Text()
		if err != nil {
			fmt.Printf("text error: %s\n", err)
		}
		r := NewResult(url, title)
		if filterProgramWithTitle(r) {
			ch <- r
		}
	}
	return nil
}

// fetchPrograms repeat fetchProgramsOn for all available progrm list pages.
func fetchPrograms(wd selenium.WebDriver, startURL string, ch chan<- Result, done chan<- bool) error {
	if err := fetchProgramsOn(wd, startURL, ch); err != nil {
		return err
	}
	curURL, err := wd.CurrentURL()
	if err != nil {
		return err
	}
	for {
		err = wd.WaitWithTimeout(WaitForXPath(ProgramNextPage), Timeout)
		if err != nil {
			return err
		}
		elem, err := wd.FindElement(selenium.ByXPATH, ProgramNextPage)
		if err != nil {
			return err
		}
		url, err := elem.GetAttribute("href")
		if err != nil {
			return err
		}
		if url == curURL {
			break
		}
		curURL = url
		fetchProgramsOn(wd, url, ch)
		time.Sleep(3 * time.Second)
	}
	done <- true
	return nil
}

func getDetailedPage(wd selenium.WebDriver, url string) (*Page, error) {
	if !strings.HasPrefix(url, "http") {
		return nil, fmt.Errorf("URL must start with http(s)\n")
	}
	if err := wd.Get(url); err != nil {
		return nil, err
	}
	if err := wd.WaitWithTimeout(WaitForXPath(DetailedPageTitle), Timeout); err != nil {
		return nil, err
	}
	elem, err := wd.FindElement(selenium.ByXPATH, DetailedPageTitle)
	if err != nil {
		return nil, err
	}
	title, err := elem.Text()
	if err != nil {
		return nil, err
	}
	elem, err = wd.FindElement(selenium.ByXPATH, DetailedPageTime)
	if err != nil {
		return nil, err
	}
	t, err := elem.Text()
	if err != nil {
		return nil, err
	}
	elem, err = wd.FindElement(selenium.ByXPATH, DetailedPageProvider)
	if err != nil {
		elem, err = wd.FindElement(selenium.ByXPATH, DetailedPageProvider2)
		if err != nil {
			return nil, err
		}
	}
	provider, err := elem.Text()
	if err != nil {
		return nil, err
	}
	p := ProviderMap[provider]
	return NewPage(url, title, p, t)
}
