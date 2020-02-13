package web

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mj37yhyy/gowb/pkg/constant"
	"github.com/mj37yhyy/gowb/pkg/model"
	"github.com/mj37yhyy/gowb/pkg/utils"
	"github.com/mj37yhyy/gowb/pkg/web/middleware"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HandlerFunc func(context.Context) (model.Response, error)
type Router struct {
	Path    string
	Method  string
	Handler HandlerFunc
}

func Bootstrap(ctx context.Context) {
	server := start(ctx)
	_signal()
	_timeout(ctx, server)
}

func start(c context.Context) *http.Server {
	conf := c.Value(constant.ConfigKey).(*utils.Config)
	routers := c.Value(constant.RoutersKey).([]Router)

	gin.SetMode(conf.Get("web.runMode").(string))

	routersInit := doRouter(c, routers)
	readTimeout := time.Minute
	writeTimeout := time.Minute
	endPoint := fmt.Sprintf(":%d", conf.Get("web.port"))
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        routersInit,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	go func() {
		// service connections
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	log.Printf("[info] start http server listening %s", endPoint)
	return server
}

func _signal() {
	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")
}

func _timeout(ctx context.Context, server *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Println("timeout of 5 seconds.")
	}
	log.Println("Server exiting")
}

func doRouter(c context.Context, routers []Router) *gin.Engine {
	return router(initGin(c), routers)
}

func initGin(c context.Context) (r *gin.Engine) {
	r = gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.Use(middleware.NoCache)
	r.Use(middleware.Options)
	r.Use(func(ctx *gin.Context) {
		ctx.Set(constant.ContextKey, c)
	})
	r.Use(middleware.RequestLogging())
	r.Use(middleware.Logger())
	return r
}

func router(r *gin.Engine, routers []Router) *gin.Engine {
	// 404 Handler.
	r.NoRoute(func(c *gin.Context) {
		c.String(http.StatusNotFound, "The incorrect API route.")
	})

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf(time.Now().String()))
	})
	for _, router := range routers {
		r.Handle(router.Method, router.Path, func(ctx *gin.Context) {
			getBody(ctx)
			getParams(ctx)
			getHeader(ctx)

			resp, err := router.Handler(getContext(ctx))

			if err != nil {
				ctx.JSON(resp.Error.Code, resp)
			} else {
				ctx.JSON(http.StatusOK, resp)
			}
		})
	}
	return r
}

func getHeader(ctx *gin.Context) {
	context.WithValue(getContext(ctx), constant.HeaderKey, ctx.Request.Header)
}

func getParams(ctx *gin.Context) {
	var params = make(map[string]string)
	for _, param := range ctx.Params {
		params[param.Key] = param.Value
	}
	context.WithValue(getContext(ctx), constant.ParamsKey, params)
	return
}

func getBody(ctx *gin.Context) {
	body, _ := ioutil.ReadAll(ctx.Request.Body)
	context.WithValue(getContext(ctx), constant.BodyKey, body)
}

func getContext(ctx *gin.Context) context.Context {
	return ctx.Value(constant.ContextKey).(context.Context)
}
