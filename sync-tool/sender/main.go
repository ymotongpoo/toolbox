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
	"flag"
	"log"
	"os"
	"time"

	"github.com/rjeczalik/notify"
	"github.com/ymotongpoo/toolbox/sync-tool"
)

var (
	fs            *flag.FlagSet
	pergeInterval *time.Duration
	secretsPath   *string
)

const PergeDuration = 1 * time.Hour

func init() {
	fs = flag.NewFlagSet("base", flag.ExitOnError)
	pergeInterval = fs.Duration("perge", DefaultPergeInterval, "perge interval duration")
	secretsPath = fs.String("secrets", synctool.DefaultSecretsFile, "path to client_secret.json")
}

func main() {
	fs.Parse(os.Args[1:])
	c := make(chan notify.EventInfo, 1)
	eventList := []notify.Event{
		notify.InCreate,
		notify.InCloseWrite,
	}
	if err := notify.Watch(".", c, eventList...); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	tick := time.NewTicker(PergeDuration)
	m := synctool.NewManager(synctool.DefaultSecretsFile) // TODO: replace file name with cli args
	err := m.Init()
	if err != nil {
		log.Fatalln(err)
	}

	for {
		select {
		case ei := <-c:
			switch {
			case ei.Event() == notify.InCloseWrite:
				log.Printf("Writing to %s is done!", ei.Path())
				go upload(m, ei)
			case ei.Event() == notify.InCreate:
				log.Printf("File %s is created!", ei.Path())
			}
		case <-tick.C:
			perge()
		}
	}
}

func upload(m *synctool.Manager, ei notify.EventInfo) {
	res, err := m.Upload(ei.Path(), "", []string{synctool.UploadTargetFolderID})
	if err != nil {
		log.Print(err)
	} else {
		log.Print(synctool.Loginfo(res))
	}
	f := synctool.NewFile(ei.Path(), res.Id)
	f.Uploaded = true
	m.AddFile(f)
}
