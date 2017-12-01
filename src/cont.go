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
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hunterhug/GoSpider/spider"
	"github.com/hunterhug/GoTool/store/myredis"
	"github.com/hunterhug/GoTool/store/mysql"
	"github.com/hunterhug/GoTool/util"
)

// 可抽离到配置文件中
const (
	// 网站
	Url  = "http://jandan.net"
	Host = "jandan.net"

	// 详情页爬虫数量
	DetailSpiderNum        = 30
	DetailSpiderNamePrefix = "detail"
	// 首页爬虫数量
	IndexSpiderNum        = 3
	IndexSpiderNamePrefix = "index"

	// 爬虫超时时间
	TimeOut = 15
	// 日志级别
	LogLevel = "info"
)

var (
	// 首页页数
	IndexPage int

	RedisClient *myredis.MyRedis

	RedisListTodo  = "jiandantodo"
	RedisListDoing = "jiandandoing"
	RedisListDone  = "jiandandone"
	IpPool         = "ipip"
	RootDir        = ""

	MysqlClient *mysql.Mysql
	RedisConfig = myredis.RedisConfig{}
	MysqlConfig = mysql.MysqlConfig{}
)

type configXX struct {
	Dir             string              `json:"dir"`
	Log             string              `json:"log"`
	DetailSpiderNum int                 `json:"detail_spider_num"`
	IndexSpiderNum  int                 `json:"index_spider_num"`
	TimeOut         int                 `json:"time_out"`
	Redis           myredis.RedisConfig `json:"redis"`
	Mysql           mysql.MysqlConfig   `json:"mysql"`
	Wait            int                 `json:"wait_time"`
}

// 设置全局
func Config(file string) {

	d, e := util.ReadfromFile(file)
	if e != nil {
		fmt.Println("config file read err:" + e.Error())
		os.Exit(-1)
	}

	xx := configXX{}
	err := json.Unmarshal(d, &xx)

	//a, _ := json.Marshal(xx)
	//fmt.Printf("%#v\n", string(a))

	if err != nil {
		fmt.Println("parse config file err:" + err.Error())
		os.Exit(-1)
	}
	// 根目录
	//RootDir = util.CurDir()
	if xx.Dir == "" {
		xx.Dir, _ = util.GetBinaryCurrentPath()
	}
	RootDir = xx.Dir

	// Redis配置
	RedisConfig = xx.Redis

	// mysql config
	MysqlConfig = xx.Mysql

	e = util.MakeDir(filepath.Join(RootDir, "data", "detail"))
	if e != nil {
		spider.Log().Panic(e.Error())
	}
	if xx.TimeOut == 0 {
		xx.TimeOut = TimeOut
	}
	spider.SetGlobalTimeout(xx.TimeOut)
	spider.SetLogLevel(LogLevel)
	indexstopchan = make(chan bool, 1)

	// 初始化爬虫们，一种多爬虫方式，设置到全局MAP中
	for i := 0; i <= IndexSpiderNum; i++ {
		s, e := spider.New(nil)
		if e != nil {
			spider.Log().Panicf("index spider %d new error: %s", i, e.Error())
		}
		// 设置随机UA
		s.SetUa(spider.RandomUa())
		s.SetWaitTime(xx.Wait) // 这里暂停方便
		spider.Pool.Set(fmt.Sprintf("%s-%d", IndexSpiderNamePrefix, i), s)
	}
	for i := 0; i <= DetailSpiderNum; i++ {
		s, e := spider.New(nil)
		if e != nil {
			spider.Log().Panicf("detail spider %d new error: %s", i, e.Error())
		}
		s.SetUa(spider.RandomUa())
		s.SetWaitTime(xx.Wait)
		spider.Pool.Set(fmt.Sprintf("%s-%d", DetailSpiderNamePrefix, i), s)
	}
}
