package main

import (
	"fmt"
	"log"
	"time"

	spider "github.com/xuanbo/colly_plus"
)

func main() {
	urls := getStartUrls()

	spider.
		// 创建爬虫
		Create().
		// 是否debug模式，默认为false
		Debug(true).
		// 限制domain，默认值'*'，即不限制domain
		Domain("*.jd.*").
		// 启动5个goroutines，默认值20
		Parallelism(5).
		// 每个请求后等待300ms，默认值0
		Sleep(300 * time.Millisecond).
		// 设置redis的连接信息，默认值如下
		RedisProperties(&spider.RedisProperties{
			// 连接地址
			Address: "127.0.0.1:6379",
			// 密码
			Password: "",
			// db
			DB: 0,
			// key前缀
			Prefix: "colly_plus",
		}).
		// 请求前回调函数，这里可以设置header等信息
		OnRequest(func(r *spider.RequestWrapper) {
			log.Printf("Visiting: %s\n", r.Request.URL)
		}).
		// 响应回调函数，这里对响应内容处理
		OnResponse(func(r *spider.ResponseWrapper, q *spider.QueueWrapper) {
			log.Printf("Visited: %s\n", r.Response.Request.URL)
			// 添加后续url到队列
			// q.Push("http://www.baidu.com")
		}).
		// 错误回调函数，处理错误信息
		OnError(func(r *spider.ResponseWrapper, err error, q *spider.QueueWrapper) {
			log.Printf("Visit: %s, went wrong: %s\n", r.Response.Request.URL, err)
		}).
		// 设置startUrls
		StartUrls(urls).
		// 运行
		Run()
}

// 京东7个产品，每个产品评价100页urls
func getStartUrls() []string {
	startUrls := make([]string, 10)
	url := "https://sclub.jd.com/comment/productPageComments.action?productId=%s&score=0&sortType=5&page=%d&pageSize=10"
	productIds := []string{"910159", "4134146", "15301899161", "5283010", "553114", "6286954", "526825"}
	for _, productId := range productIds {
		for i := 0; i < 100; i++ {
			startUrls = append(startUrls, fmt.Sprintf(url, productId, i))
		}
	}
	return startUrls
}
