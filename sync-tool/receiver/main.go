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
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"google.golang.org/api/drive/v3"
)

const (
	UploadTargetFolderId = "1QaF-81k04ieUk4RB97PU1eALP0JnXN4S"
	EncodeDoneFolderId   = "1i0GSCuF10lW1sx3A_vDGbvjAKIxPS2yM"
	PoleInterval         = 10 * time.Minute
)

func main() {
	t := time.NewTicker(PollInterval)
	for {
		select {
		case c := <-t.C:
			log.Println(c)
		}
	}
}
