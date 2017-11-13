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
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/knq/chromedp"
	"github.com/knq/chromedp/cdp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := chromedp.New(ctx, chromedp.WithLog(log.Printf))
	if err != nil {
		log.Fatalln(err)
	}
	var site, res string
	err = c.Run(ctx, googleSearch("site:brank.as", "Home", &site, &res))
	if err != nil {
		log.Fatalln(err)
	}
	err = c.Shutdown(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	err = c.Wait()
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("saved screenshot from search result listing `%s` (%s)\n", res, site)
}

func googleSearch(q, text string, site, res *string) chromedp.Tasks {
	var buf []byte
	sel := fmt.Sprintf(`//a[text()[contains(., '%s')]]`, text)
	return chromedp.Tasks{
		chromedp.Navigate(`https://www.google.com/`),
		chromedp.WaitVisible(`#hplogo`, chromedp.ByID),
		chromedp.SendKeys(`#lst-ib`, q+"\n", chromedp.ByID),
		chromedp.WaitVisible(`#res`, chromedp.ByID),
		chromedp.Text(sel, res),
		chromedp.Click(sel),
		chromedp.WaitVisible(`a[href="/brankas-for-business"]`, chromedp.ByQuery),
		chromedp.WaitNotVisible(`.preloader-content`, chromedp.ByQuery),
		chromedp.Location(site),
		chromedp.ScrollIntoView(`.banner-section.third-section`, chromedp.ByQuery),
		chromedp.Sleep(2 * time.Second), // wait for animation to finish
		chromedp.Screenshot(`.banner-section.third-section`, &buf, chromedp.ByQuery),
		chromedp.ActionFunc(func(context.Context, cdp.Handler) error {
			return ioutil.WriteFile("screenshot.png", buf, 0644)
		}),
	}
}
