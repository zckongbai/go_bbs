package middlewares

import (
	"github.com/fpay/gopress"
)

// NewSessionMiddleware returns session middleware.
func NewSessionMiddleware() gopress.MiddlewareFunc {
	return func(next gopress.HandlerFunc) gopress.HandlerFunc {
		return func(c gopress.Context) error {
			// Uncomment this line if this middleware requires accessing to services.
			// services := gopress.AppFromContext(c).Services()

			//store, err := session.NewRedisStore(32, "tcp", "localhost:6379", "", []byte("secret"))
			//if err != nil {
			//	panic(err)
			//}

			//c.app().Use(session.Middleware(sessions.NewCookieStore([]byte("secret"))))
			return next(c)
		}
	}
}
