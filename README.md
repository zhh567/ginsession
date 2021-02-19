# ginsession

为 gin 框架添加 session 功能的中间件。可选择使用内存或 redis 。

```go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	ginsession "github.com/zhh567/gin-session"
	"log"
)

// 一个例子，在 / 通过url接收参数；在 /list 显示。
func main() {
	r := gin.Default()

	sessionMgr, err := ginsession.CreateSessionMgr("memory", "")
	//sessionMgr, err := ginsession.CreateSessionMgr("redis", "127.0.0.1:6379")
	if err != nil {
		fmt.Println(err)
		return
	}
	r.Use(ginsession.SessionMiddleware(
		sessionMgr,
		ginsession.CookieOptions{
			Path:     "/",
			Domain:   "127.0.0.1",
			MaxAge:   60,
			Secure:   false,
			HttpOnly: true,
		},
	))
	r.GET("/", func(c *gin.Context) {
		value := c.Query("key")
		data, ok := c.Get("session")
		if ok {
			session, ok := data.(ginsession.Session)
			if ok {
				session.Set("key", value)
			} else {
				log.Println("assert session failed")
			}
		} else {
			log.Println("get session failed")
		}
	})

	r.GET("/list", func(c *gin.Context) {
		data, ok := c.Get("session")
		if ok {
			session, ok := data.(ginsession.Session)
			if ok {
				v, err := session.Get("key")
				if err == nil {
					c.String(200, fmt.Sprintf("%s", v))
				}
			} else {
				log.Println("assert session failed")
			}
		} else {
			log.Println("get session failed")
		}
	})

	if err := r.Run("127.0.0.1:80"); err != nil {
		fmt.Println(err)
	}
}
```