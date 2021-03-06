/*
Copyright 2017 by GoSpider author.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License
*/
package src

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/hunterhug/marmot/expert"
	"github.com/hunterhug/parrot/util"
)

// 解析页面数量
// 此策略已经更改, 弃用!!!
func ParseIndexNum(data []byte) error {
	doc, e := expert.QueryBytes(data)
	if e != nil {
		return e
	}
	s := doc.Find(".pages").Text()
	temp := strings.Split(s, "/")
	if len(temp) != 2 {
		return errors.New("index page not found")
	}
	result := strings.Replace(strings.TrimSpace(temp[1]), ",", "", -1)
	i, e := util.SI(result)
	if e != nil {
		return e
	}
	IndexPage = i
	return nil
}

// 提取信息
func ParseIndex(data []byte) []string {
	if strings.Contains(string(data), "too many requests") {
		fmt.Println("速度过快, 待引入IP池, 见:ip.go")
		os.Exit(1)
	}
	list := []string{}
	doc, e := expert.QueryBytes(data)
	if e != nil {
		return list
	}

	doc.Find(".post").Each(func(num int, node *goquery.Selection) {
		//title := node.Find("h2").Text()
		//if title == "" {
		//	return
		//}
		url, _ := node.Find("h2").Find("a").Attr("href")
		if url == "" {
			return
		}
		//tag := node.Find(".time_s").Text()
		//if strings.Contains(tag, "·") {
		//	tag = strings.Split(tag, "·")[1]
		//}
		//fmt.Printf("%s,%s,%s\n", title, url, tag)
		list = append(list, url)
	})
	return list
}

func ParseDetail(data []byte) map[string]string {
	returnmap := map[string]string{
		"title": "", "tags": "", "content": "", "shortcontent": "",
	}
	doc, e := expert.QueryBytes(data)
	if e != nil {
		return returnmap
	}
	// 标题
	title := doc.Find("title").Text()
	if strings.TrimSpace(title) == "" {
		return returnmap
	}
	shortcontent, _ := doc.Find(`meta[name="description"]`).Attr("content")
	tags, _ := doc.Find(`meta[name="keywords"]`).Attr("content")

	result := ""
	doc.Find("#content").Find(".post p").Each(func(num int, node *goquery.Selection) {
		temp, _ := node.Html()
		result = result + "<p>" + temp + "</p>"
	})

	returnmap["title"] = strings.Replace(title, "\"", "'", -1)
	returnmap["tags"] = strings.Replace(tags, "\"", "'", -1)
	returnmap["shortcontent"] = strings.Replace(shortcontent, "\"", "'", -1)
	returnmap["content"] = strings.Replace(result, "\"", "'", -1)
	return returnmap
}
