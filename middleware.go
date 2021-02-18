package ginsession

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type CookieOptions struct {
	Path     string
	Domain   string
	MaxAge   int
	Secure   bool
	HttpOnly bool
}

const (
	// sessionName is name in cookie
	sessionName = "session_id"

	// SessionContextName used to transfer in context
	SessionContextName = "session"
)

// SessionMiddleware set session for every request
func SessionMiddleware(sm SessionMgr, options CookieOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		var session Session
		sessionID, err := c.Cookie(sessionName)
		if err != nil {
			session = sm.CreateSession()
			sessionID = session.ID()
		} else {
			session, err = sm.GetSession(sessionID)
			if err != nil {
				session = sm.CreateSession()
				sessionID = session.ID()
			}
		}
		session.SetExpired(options.MaxAge)
		c.Set(SessionContextName, session)
		c.SetCookie(sessionName, sessionID, options.MaxAge,
			options.Path, options.Domain, options.Secure, options.HttpOnly)
		defer func() {
			if session.IsRedis() {
				sm.Clear(sessionID)
			}
		}()

		c.Next()
	}
}

// AuthMiddleware used to detect whether user is logged in.
// Must used after SessionMiddleware in chain.
// Detect log status by variable named "isLogin" in session.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		v, ok := c.Get(SessionContextName)
		if ok {
			session, ok := v.(Session)
			if ok {
				isLogin, err := session.Get("isLogin")
				if err == nil {
					if isLogin.(bool) == true {
						c.Next()
					}
				}
			}
		}
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
