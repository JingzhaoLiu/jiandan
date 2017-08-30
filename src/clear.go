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

import "github.com/hunterhug/GoSpider/spider"

// 将Doing移到Todo
func Clear() {
	for {
		s, _ := RedisClient.Brpoplpush(RedisListDoing, RedisListTodo, 0)
		spider.Log().Info("movw :" + s)
	}
}