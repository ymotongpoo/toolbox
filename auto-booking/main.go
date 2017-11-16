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
	"log"
	"time"

	"github.com/knq/chromedp"
	"github.com/knq/chromedp/cdp"
)


func main() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c, err := chromedp.New(ctx, chromedp.WithErrorf(log.Printf))
	if err != nil {
		log.Fatal(err)
	}
	res, err := getTitle(ctx, c)
	if err != nil {
		log.Fatalf("could not list awesome go projects: %v", err)
	}
	err = c.Shutdown(ctx)
	if err != nil {
		log.Fatal(err)
	}
	err = c.Wait()
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res)
}

func getTitle(ctx context.Context, c *chromedp.CDP) (string, error) {
	var cancel func()
	ctx, cancel = context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	if err := c.Run(ctx, chromedp.Navigate(`http://kh31n.hatenablog.jp/entry/2017/04/09/172247`)); err != nil {
		return "", fmt.Errorf("could not navigate to github: %v", err)
	}
	if err := c.Run(ctx, chromedp.WaitVisible(`//*[@id="entry-10328749687235837314"]/div/header/h1/a`)); err != nil {
		return "", fmt.Errorf("could not get section: %v", err)
	}

	// get project link text
	var projects []*cdp.Node
	if err := c.Run(ctx, chromedp.Nodes(`//*[@id="entry-10328749687235837314"]/div/header/h1/a/text()`, &projects)); err != nil {
		return "", fmt.Errorf("could not get projects: %v", err)
	}

	return projects[0].NodeValue, nil
}
