package middleware

import (
	"time"

	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v5"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"readwillbe/internal/cache"
	"readwillbe/internal/model"
	"readwillbe/internal/repository"
)

func UserMiddleware(db *gorm.DB, userCache *cache.UserCache, cfg model.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			sess, err := session.Get(SessionKey, c)
			if err != nil {
				logrus.Warnf("Failed to get session: %v", err)
				return next(c)
			}
			if sess.Values[SessionUserIDKey] != nil {
				userID, ok := sess.Values[SessionUserIDKey].(uint)
				if !ok {
					return next(c)
				}

				user, found := userCache.Get(userID)
				if !found {
					var err error
					user, err = repository.GetUserByID(db.WithContext(c.Request().Context()), userID)
					if err != nil {
						if errors.Is(err, gorm.ErrRecordNotFound) {
							delete(sess.Values, SessionUserIDKey)
							_ = sess.Save(c.Request(), c.Response())
							return next(c)
						}
						return errors.Wrap(err, "getting user by id")
					}
					userCache.Set(user)
				}

				c.Set(UserKey, user)

				shouldSave := false

				if sess.Values[SessionUserIDKey] != user.ID {
					sess.Values[SessionUserIDKey] = user.ID
					shouldSave = true
				}

				lastSeen, ok := sess.Values[SessionLastSeenKey].(int64)
				now := time.Now().Unix()
				if !ok || now-lastSeen > SessionRefreshInterval {
					sess.Values[SessionLastSeenKey] = now
					shouldSave = true
				}

				if shouldSave {
					sess.Options = GetSecureSessionOptions(cfg)
					if err := sess.Save(c.Request(), c.Response()); err != nil {
						return errors.Wrap(err, "saving session")
					}
				}
			}
			return next(c)
		}
	}
}

func GetSessionUser(c *echo.Context) (model.User, bool) {
	u := c.Get(UserKey)
	if u != nil {
		user, ok := u.(model.User)
		if !ok {
			return model.User{}, false
		}
		logrus.Debugf("Found session user ID=%d", user.ID)
		return user, true
	}
	return model.User{}, false
}
