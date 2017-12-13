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
	"fmt"
	"path/filepath"

	"github.com/hunterhug/marmot/miner"
	"github.com/hunterhug/parrot/util"
)

var (
	// 信号量
	indexstopchan chan bool
	maxpage       = 100000 // 预估的最大页数的两倍
	pageSpider    *miner.Worker
)

func init() {
	pageSpider, _ = miner.New(nil) // 预估页数的爬虫
}

// 首页启动入口，包括所有非详情页面的抓取
// 抓取网址到redis，因为页数经常变动，所以这个爬虫比较暴力，借助文件夹功能接力，如果页面更新，请将data数据夹删除
func IndexSpiderRun() {
	// 获取首页页数并把首页网址打到redis
	IndexStep()
	// 按顺序抓取页面，打到redis
	PagesStep()
}

// 步骤1：首页随便取只爬虫抓取
func IndexStep() {
	s, ok := miner.Pool.Get(IndexSpiderNamePrefix + "-0")
	if !ok {
		miner.Log().Panic("IndexStep:Get Index Spider error!")
	}
	// 爬取首页
	s.SetUrl(Url).SetMethod("get")
	data, e := s.Go()
	if e != nil {
		// 错误直接退出
		miner.Log().Panicf("Get Index Error:%s", e.Error())
	}

	miner.Log().Info("Catch Index!")

	// 实验的
	indexfile := filepath.Join(RootDir, "data", "index.html")
	e = util.SaveToFile(indexfile, data)
	if e != nil {
		miner.Log().Errorf("Save Index Error:%s", e.Error())
	}

	// 获取页数
	// e = ParseIndexNum(data)
	// if e != nil {
	// 	miner.Log().Panic(e.Error())
	// }
	pagefirst := 1      // 1
	pagelast := maxpage // 100
	for {
		pagenow := (pagefirst + pagelast) / 2 // 50  1-50 50-100
		miner.Log().Info(pagefirst, pagenow, pagelast)
		if pagenow == pagefirst {
			break
		}

		pageSpider.SetUrl(fmt.Sprintf("%s/page/%d", Url, pagenow))
		result, e := pageSpider.Get()
		if e != nil {
			miner.Log().Panic(e.Error())
		}

		if len(ParseIndex(result)) == 0 {
			pagelast = pagenow // 1 50
		} else {
			pagefirst = pagenow // 50 100
		}
	}

	miner.Log().Info(pagefirst)
	// os.Exit(1)
	IndexPage = pagefirst

	SentRedis(ParseIndex(data))
}

// 步骤2：分配任务
func PagesStep() {
	urllist := []string{}
	for i := 2; i <= IndexPage; i++ {
		urllist = append(urllist, fmt.Sprintf("%s/page/%d", Url, i))
	}
	// 分配任务
	tasks, e := util.DevideStringList(urllist, IndexSpiderNum)
	if e != nil {
		miner.Log().Panic(e.Error())
	}
	// 任务开始
	for i, task := range tasks {
		go PagesTaskGoStep(i, task)
	}
	for i, _ := range tasks {
		// 等待爬虫结束
		<-indexstopchan
		miner.Log().Infof("index miner %s-%d finish", IndexSpiderNamePrefix, i)
	}
}

// 步骤2接力：多只爬虫并发抓页面
func PagesTaskGoStep(name int, task []string) {
	var e error
	var data []byte
	// 获取池中爬虫
	spidername := fmt.Sprintf("%s-%d", IndexSpiderNamePrefix, name)
	s, ok := miner.Pool.Get(spidername)
	if !ok {
		miner.Log().Panicf("Pool Spider %s not get", spidername)
	}
Outloop:
	for _, url := range task {
		// 文件存在，那么不抓
		pagename := fmt.Sprintf("%s.html", util.ValidFileName(url))
		savepath := filepath.Join(RootDir, "data", pagename)
		if util.FileExist(savepath) {
			miner.Log().Infof("page %s Exist", pagename)
			data, e = util.ReadfromFile(savepath)
			if e != nil {
				miner.Log().Errorf("take data from exist file error:%s", e.Error())
			} else {
				SentRedis(ParseIndex(data))
			}
			continue
		}
		s.SetUrl(url)
		s.SetRefer(s.Preurl)
		retrynum := 5
		for {
			if retrynum == 0 {
				goto Outloop
			}
			data, e = s.Go()
			if e != nil {
				miner.Log().Errorf("%s: index page %s fetch error:%s,remain %d times", spidername, url, e.Error(), retrynum)
				retrynum = retrynum - 1
				continue
			}
			SentRedis(ParseIndex(data))
			miner.Log().Infof("%s:index page %s fetch!", spidername, url)
			break
		}

		// 保存文件
		e = util.SaveToFile(savepath, data)
		if e != nil {
			miner.Log().Errorf("Save page %s Fail:%s", pagename, e.Error())
		}
		miner.Log().Infof("Save page %s Done", pagename)
	}

	indexstopchan <- true
}
