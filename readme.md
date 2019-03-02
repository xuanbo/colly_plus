# CollyPlus

> A functional spider based on [colly](https://github.com/gocolly/colly).

## Features

* Functional
* Redis Queue

## Example

```go
package main

import (
	"log"
	"time"

	spider "github.com/xuanbo/colly_plus"
)

func main() {
	spider.Create().
		Parallelism(5).
		Sleep(300 * time.Millisecond).
		OnResponse(func(r *spider.ResponseWrapper, q *spider.QueueWrapper) {
			resp := r.Response
			log.Printf("Visited: %s, body: %s.\n", resp.Request.URL, resp.Body)
		}).
		OnError(func(r *spider.ResponseWrapper, err error, q *spider.QueueWrapper) {
			log.Printf("Visit: %s, went wrong: %s\n", r.Response.Request.URL, err)
		}).
		StartUrl("https://github.com/xuanbo/colly_plus").
		Run()
}

```

See [examples](./examples) folder for more detailed examples.

## Installation

```
go get github.com/xuanbo/colly_plus
```