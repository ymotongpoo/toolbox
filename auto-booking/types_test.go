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
	"reflect"
	"testing"
	"time"
)

func Test_parseTime(t *testing.T) {
	type result struct {
		s time.Time
		e time.Time
	}
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	in := []string{
		"2017年11月18日（土）  25時20分～26時10分",
		"2017年11月19日（日）  23時10分～24時05分",
		"2017年11月19日（日）  8時15分～9時00分",
	}
	want := []result{
		{time.Date(2017, 11, 19, 1, 20, 0, 0, jst), time.Date(2017, 11, 19, 2, 10, 0, 0, jst)},
		{time.Date(2017, 11, 19, 23, 10, 0, 0, jst), time.Date(2017, 11, 20, 0, 5, 0, 0, jst)},
		{time.Date(2017, 11, 19, 8, 15, 0, 0, jst), time.Date(2017, 11, 19, 9, 0, 0, 0, jst)},
	}

	for i, p := range in {
		s, e, err := parseTime(p)
		if err != nil {
			t.Fatalf("error: %s", err)
		}
		r := result{s, e}
		if !reflect.DeepEqual(want[i], r) {
			t.Fatalf("want: %s & %s, out: %s & %s", want[i].s, want[i].s, r.s, r.e)
		}
	}
}
func Test_parseAtID(t *testing.T) {
	in := []string{
		"job 1 at Sun Nov 19 14:25:00 2017",
		"job 24 at Sun Nov 19 21:06:00 2017",
		"job 10773 at Fri Nov 24 14:20:00 2017",
	}
	want := []int{
		1,
		24,
		10773,
	}
	for i, r := range in {
		id := parseAtID(r)
		if id != want[i] {
			t.Fatalf("want: %d, out: %d", want[i], id)
		}
	}
}
