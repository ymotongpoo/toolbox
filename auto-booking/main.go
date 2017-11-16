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
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx, subcancel := context.WithTimeout(ctx, 25*time.Second)
	defer subcancel()
	c, err := chromedp.New(ctx, chromedp.WithErrorf(log.Printf))
	if err != nil {
		log.Fatalln(err)
	}
	nodes, err := nodeValueTest(ctx, c)
	if err != nil {
		log.Fatalln(err)
	}
	var text string
	for i, n := range nodes {
		c.Run(ctx, chromedp.Text(n, &text))
		fmt.Println(i, text)
	}
	err = c.Shutdown(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	err = c.Wait()
	if err != nil {
		log.Fatalln(err)
	}
}
