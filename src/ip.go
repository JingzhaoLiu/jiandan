package src

import (
	"fmt"

	"github.com/hunterhug/GoSpider/spider"
)

func GetIP() string {
	spider.Logger.Debug("Get IP...")
	iptemp, ierr := RedisClient.Brpop(0, IpPool)
	// ip null return,maybe forever not happen
	if ierr != nil {
		panic("ip:" + ierr.Error())
	}
	ip := iptemp[1]
	spider.Logger.Debug("Get IP done:" + ip)
	return ip
}

func Sentiptoredis(ips []string) string {
	if len(ips) == 0 {
		return "IP Empty"
	}
	returns := ""
	for _, ip := range ips {
		_, err := RedisClient.Lpush(IpPool, ip)
		if err != nil {
			fmt.Printf("%s error:%v\n", ip, err)
			returns = returns + fmt.Sprintf("%s error:%v\n", ip, err)
		} else {
			fmt.Printf("%s success\n", ip)
			returns = returns + fmt.Sprintf("%s success\n", ip)
		}
	}
	return returns
}
