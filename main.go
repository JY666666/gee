package main

import (
	"gee"
	"net/http"
)

func main() {
	r := gee.New()

	r.GET("/", func(c *gee.Context) {
		c.Html(http.StatusOK, "<h1>hello Gee<h1>")
	})

	r.GET("/hello", func(c *gee.Context) {
		c.String(http.StatusOK, "hello %s,you're at %s \n", c.Query("name"), c.Path)
	})

	r.Run(":9999")
}
