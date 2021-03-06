package recovery_lite

import (
	"net/http"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/ldeng7/ginx"
	"github.com/ldeng7/go-logger-lite/logger"
)

func recovery(gc *gin.Context, logger *logger.Logger, depth int, callback func(*gin.Context, interface{})) {
	p := recover()
	if nil == p {
		return
	}

	if nil != callback {
		callback(gc, p)
	}
	logger.Err("panic: ", p)
	for i := 2; i < depth+2; i++ {
		_, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		logger.Errf("panic on %s:%d", file, line)
	}

	c := ginx.Context{gc}
	c.RenderError(&ginx.RespError{Status: http.StatusInternalServerError})
	gc.Abort()
}

func Recovery(logger *logger.Logger, depth int, callback func(*gin.Context, interface{})) gin.HandlerFunc {
	return func(gc *gin.Context) {
		defer recovery(gc, logger, depth, callback)
		gc.Next()
	}
}
