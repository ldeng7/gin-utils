package redis_cookie_auth

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis"
	"github.com/ldeng7/ginx"
	"github.com/ldeng7/ginx/middlewares/redis_auth"
)

const RED_KEY_PREF = "cauth:"
const GIN_META_UID = "cauth:uid"
const COOKIE_NAME_UID = "uid"
const COOKIE_NAME_TOKEN = "token"

type RedisCookieAuth struct {
	*redis_auth.RedisAuth
}

func init() {
	rand.Seed(time.Now().Unix())
}

func New(red *redis.Client, namespace string) *RedisCookieAuth {
	return &RedisCookieAuth{
		redis_auth.New(red, namespace, RED_KEY_PREF),
	}
}

func (a *RedisCookieAuth) auth(gc *gin.Context) (int, string, error) {
	cookieUid, err := gc.Request.Cookie(COOKIE_NAME_UID)
	if nil != err {
		return http.StatusUnauthorized, err
	}
	cookieToken, err := gc.Request.Cookie(COOKIE_NAME_TOKEN)
	if nil != err {
		return http.StatusUnauthorized, err
	}
	status, err := a.Read(cookieUid.Value, cookieToken.Value)
	return status, cookieUid.Value, err
}

func (a *RedisCookieAuth) Middleware() gin.HandlerFunc {
	return func(gc *gin.Context) {
		status, uid, err := a.auth(gc)
		if http.StatusOK != status {
			c := ginx.Context{gc}
			c.RenderError(&ginx.RespError{StatusCode: status, Message: err.Error()})
			gc.Abort()
			return
		}
		gc.Set(GIN_META_UID, uid)
		gc.Next()
	}
}

func (a *RedisCookieAuth) Set(gc *gin.Context, uid string, ttl time.Duration) error {
	now := time.Now()
	h := md5.New()
	h.Write([]byte(now.String()))
	h.Write([]byte(strconv.Itoa(rand.Int())))
	token := fmt.Sprintf("%x", h.Sum(nil))

	if err := a.Write(uid, token, ttl); nil != err {
		return err
	}
	expire := now.Add(ttl)
	maxAge := int(ttl / time.Second)
	http.SetCookie(gc.Writer, &http.Cookie{
		Name:    COOKIE_NAME_UID,
		Value:   uid,
		Path:    "/",
		Expires: expire,
		MaxAge:  maxAge,
	})
	http.SetCookie(gc.Writer, &http.Cookie{
		Name:    COOKIE_NAME_TOKEN,
		Value:   token,
		Path:    "/",
		Expires: expire,
		MaxAge:  maxAge,
	})
	return nil
}
