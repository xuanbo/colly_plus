package main

import (
	"log"
	"time"

	spider "github.com/xuanbo/colly_plus"
)

func main() {
	spider.
		// 创建爬虫
		Create().
		// 启动5个goroutines，默认值20
		Parallelism(5).
		// 每个请求后等待300ms，默认值0
		Sleep(300 * time.Millisecond).
		// 响应回调函数，这里对响应内容处理
		OnResponse(func(r *spider.ResponseWrapper, q *spider.QueueWrapper) {
			resp := r.Response
			log.Printf("Visited: %s, body: %s.\n", resp.Request.URL, resp.Body)
		}).
		// 错误回调函数，处理错误信息
		OnError(func(r *spider.ResponseWrapper, err error, q *spider.QueueWrapper) {
			log.Printf("Visit: %s, went wrong: %s\n", r.Response.Request.URL, err)
		}).
		// 设置startUrl
		StartUrl("https://github.com/xuanbo").
		// 运行
		Run()
}
