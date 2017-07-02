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
package main

import (
	//"fmt"
	"github.com/hunterhug/jiandan"
	"os"
	"os/signal"
)

var Clear = false

func main() {
	if Clear {
		// Reids中Doing的迁移到Todo，需手动，var Clear = true
		go jiandan.Clear()
	} else {
		// 首页爬虫爬取
		go jiandan.IndexSpiderRun()

		// 详情页抓取
		go jiandan.DetailSpidersRun()
	}

	c := make(chan os.Signal)
	//监听指定信号
	signal.Notify(c, os.Interrupt)

	//阻塞直至有信号传入
	<-c
}
