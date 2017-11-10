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

	"github.com/rjeczalik/notify"
	"github.com/ymotongpoo/toolbox/sync-tool"
)

func main() {
	c := make(chan notify.EventInfo, 1)

	eventList := []notify.Event{
		notify.InCreate,
		notify.InCloseWrite,
	}
	if err := notify.Watch(".", c, eventList...); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	m := synctool.NewManager(synctool.DefaultSecretsFile) // TODO: replace file name with cli args
	err := m.Init()
	if err != nil {
		log.Fatalln(err)
	}

	for {
		switch ei := <-c; ei.Event() {
		case notify.InCloseWrite:
			log.Printf("Writing to %s is done!", ei.Path())
			go upload(m)
		case notify.InCreate:
			log.Printf("File %s is created!", ei.Path())
		}
	}
}

func upload(m *synctool.Manager) {
	res, err := m.Upload(ei.Path(), "", []string{synctool.UploadTargetFolderID})
	if err != nil {
		log.Print(err)
	} else {
		log.Print(synctool.Loginfo(res))
	}
}
