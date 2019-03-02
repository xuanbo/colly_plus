package colly_plus

import (
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/gocolly/colly/queue"
	"github.com/gocolly/redisstorage"
)

// 爬虫
type Spider struct {
	debug            bool
	domain           string
	parallelism      int
	delay            time.Duration
	redisProperties  *RedisProperties
	startUrls        []string
	requestCallback  RequestCallback
	responseCallback ResponseCallback
	errorCallback    ErrorCallback
	collector        *colly.Collector
	storage          *redisstorage.Storage
	queue            *QueueWrapper
}

// redis配置
type RedisProperties struct {
	Address  string
	Password string
	DB       int
	Prefix   string
}

// colly队列包装
type QueueWrapper struct {
	q *queue.Queue
}

// colly请求包装
type RequestWrapper struct {
	Request *colly.Request
}

// colly响应包装
type ResponseWrapper struct {
	Response *colly.Response
}

// 请求前回调函数
type RequestCallback func(*RequestWrapper)

// 响应回调函数
type ResponseCallback func(*ResponseWrapper, *QueueWrapper)

// 错误回调函数
type ErrorCallback func(*ResponseWrapper, error, *QueueWrapper)

// 创建Spider
func Create() *Spider {
	return &Spider{
		domain:      "*",
		parallelism: 20,
		redisProperties: &RedisProperties{
			Address:  "127.0.0.1:6379",
			Password: "",
			DB:       0,
			Prefix:   "colly_plus",
		},
		startUrls: make([]string, 10),
	}
}

// 设置debug
func (s *Spider) Debug(debug bool) *Spider {
	s.debug = debug
	return s
}

// 设置domain限制
func (s *Spider) Domain(domain string) *Spider {
	s.domain = domain
	return s
}

// 设置并行度
func (s *Spider) Parallelism(num int) *Spider {
	s.parallelism = num
	return s
}

// 设置请求后等待时间
func (s *Spider) Sleep(duration time.Duration) *Spider {
	s.delay = duration
	return s
}

// 设置redis配置
func (s *Spider) RedisProperties(properties *RedisProperties) *Spider {
	s.redisProperties = properties
	return s
}

// 添加startUrl
func (s *Spider) StartUrl(url string) *Spider {
	s.startUrls = append(s.startUrls, url)
	return s
}

// 添加startUrls
func (s *Spider) StartUrls(urls []string) *Spider {
	s.startUrls = append(s.startUrls, urls...)
	return s
}

// 请求前回调函数
func (s *Spider) OnRequest(callback RequestCallback) *Spider {
	s.requestCallback = callback
	return s
}

// 响应回调函数
func (s *Spider) OnResponse(callback ResponseCallback) *Spider {
	s.responseCallback = callback
	return s
}

// 错误回调函数
func (s *Spider) OnError(callback ErrorCallback) *Spider {
	s.errorCallback = callback
	return s
}

// 启动
func (s *Spider) Run() {
	// 初始化colly的collector
	if s.debug {
		s.collector = colly.NewCollector(colly.Debugger(&debug.LogDebugger{}))
	} else {
		s.collector = colly.NewCollector()
	}

	// 限制colly的domain、goroutines个数、请求延迟时间
	err := s.collector.Limit(&colly.LimitRule{
		DomainGlob:  s.domain,
		Parallelism: s.parallelism,
		Delay:       s.delay,
	})
	checkError(err)

	// 创建redis storage
	s.storage = &redisstorage.Storage{
		Address:  s.redisProperties.Address,
		Password: s.redisProperties.Password,
		DB:       s.redisProperties.DB,
		Prefix:   s.redisProperties.Prefix,
	}

	// 添加storage到collector
	err = s.collector.SetStorage(s.storage)
	checkError(err)

	// 关闭storage，释放资源
	defer func() {
		err = s.storage.Client.Close()
		checkError(err)
	}()

	// 创建队列，采用redis storage
	q, err := queue.New(s.parallelism, s.storage)
	s.queue = &QueueWrapper{q}
	checkError(err)

	// 将startUrls放入队列
	s.queue.PushMulti(s.startUrls)

	// 设置callback
	s.collector.OnRequest(func(r *colly.Request) {
		if s.requestCallback != nil {
			s.requestCallback(&RequestWrapper{r})
		}
	})
	s.collector.OnResponse(func(r *colly.Response) {
		if s.responseCallback != nil {
			s.responseCallback(&ResponseWrapper{r}, s.queue)
		}
	})
	s.collector.OnError(func(r *colly.Response, err error) {
		if s.errorCallback != nil {
			s.errorCallback(&ResponseWrapper{r}, err, s.queue)
		}
	})

	// 启动，开始消费队列中的urls
	err = s.queue.q.Run(s.collector)
	checkError(err)
}

// 添加url到队列
func (qr *QueueWrapper) Push(url string) {
	err := qr.q.AddURL(url)
	checkError(err)
}

// 添加urls到队列
func (qr *QueueWrapper) PushMulti(urls []string) {
	for _, url := range urls {
		qr.Push(url)
	}
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}
