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

// 详情页爬虫
func DetailSpidersRun() {
	for i := 0; i < DetailSpiderNum; i++ {
		go DetailTaskStep(i)
	}
}

func DetailTaskStep(name int) {
	spidername := fmt.Sprintf("%s-%d", DetailSpiderNamePrefix, name)
	detailpath := filepath.Join(RootDir, "data", "detail")
	s, ok := miner.Pool.Get(spidername)
	if !ok {
		miner.Log().Panicf("Pool Spider %s not get", spidername)
	}

	for {
		// 将Todo移到Doing
		url, e := RedisClient.Brpoplpush(RedisListTodo, RedisListDoing, 0)
		if e != nil {
			miner.Log().Errorf("BrpopLpush %s error:%s", url, e.Error())
			break
		}
		// Done已经存在
		ok, _ := RedisClient.Hexists(RedisListDone, url)
		if ok {
			// 删除Doing!
			RedisClient.Lrem(RedisListDoing, 0, url)
			continue
		}
		// 文件存在不抓！
		filename := filepath.Join(detailpath, util.ValidFileName(url))
		if util.FileExist(filename) {
			miner.Log().Infof("file:%s exist", filename)
			// 删除Doing!
			RedisClient.Lrem(RedisListDoing, 0, url)
			// 读取后解析存储
			data, e := util.ReadfromFile(filename)
			if e != nil {
				miner.Log().Errorf("take from file %s error: %s", filename, e.Error())
			} else {
				SaveToMysql(url, ParseDetail(data))
			}
			RedisClient.Hset(RedisListDone, url, "")
			continue
		}
		s.SetUrl(url)
		retrynum := 5
		for {
			if retrynum == 0 {
				break
			}
			data, e := s.Go()
			if e != nil {
				miner.Log().Errorf("%s:detail url %s catch error:%s remian %d times", spidername, url, e.Error(), retrynum)
				retrynum = retrynum - 1
				continue
			} else {
				miner.Log().Infof("catch url:%s", url)
				e := util.SaveToFile(filename, data)
				if e != nil {
					miner.Log().Errorf("file %s save error:%s", filename, e.Error())
				}

				SaveToMysql(url, ParseDetail(data))

				// 删除Doing!
				RedisClient.Lrem(RedisListDoing, 0, url)
				// 送到Done中
				RedisClient.Hset(RedisListDone, url, "")
				break
			}
		}
	}
}
