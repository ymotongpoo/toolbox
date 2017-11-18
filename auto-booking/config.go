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

import "regexp"

func programTitleFilter() []*regexp.Regexp {
	ptn := []string{
		`.*花子とアン.*`,
		`.*わろてんか.*`,
		`.*デザインあ.*`,
		`.*ピタゴラスイッチ.*`,
		`.*地球ドラマチック.*`,
		`.*BS世界のドキュメンタリー.*`,
		`.*クローズアップ現代.*`,
		`.*ねほりんぱほりん.*`,
		`.*旅するスペイン語.*`,
		`.*タモリ倶楽部.*`,
		`.*ブラタモリ.*`,
		`.*鉄腕.*`,
		`.*ドキュメント72時間.*`,
		`.*ルパン三世.*`,
		`.*ゴッドタン.*`,
		`.*3月のライオン.*`,
		`.*プロフェッショナル　仕事の流儀.*`,
		`.*オイコノミア.*`,
		`.*日本の話芸.*`,
		`.*所さんの目がテン.*`,
		`.*NHKスペシャル.*`,
		`.*世界ふしぎ発見.*`,
		`.*探偵！ナイトスクープ.*`,
		`.*ぼくらはマンガで強くなった.*`,
		`.*BS1スペシャル.*`,
		`.*世界仰天ニュース.*`,
		`.*水曜どうでしょう.*`,
		`.*新日本風土記.*`,
		`.*将棋.*`,
		`.*探検バクモン.*`,
		`.*日本の話芸.*`,
		`.*落語研究会.*`,
		`.*超入門！落語　THE　MOVIE.*`,
		`.*超AI入門.*`,
		`.*世界史.*`,
		`.*日本史.*`,
		`.*地理.*`,
		`.*ビジネス基礎.*`,
		`.*家庭総合.*`,
		`.*社会と情報.*`,
		`.*簿記.*`,
		`.*ATPテニス.*`,
		`.*ウィンブルドン.*`,
		`.*マツコの知らない世界.*`,
		`.*Why！？プログラミング.*`,
	}
	ret := []*regexp.Regexp{}
	for _, p := range ptn {
		ret = append(ret, regexp.MustCompile(p))
	}
	return ret
}
