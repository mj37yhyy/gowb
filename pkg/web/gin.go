package web

import (
	"context"
	"fmt"
	"github.com/chenjiandongx/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/mj37yhyy/gowb/pkg/config"
	"github.com/mj37yhyy/gowb/pkg/constant"
	"github.com/mj37yhyy/gowb/pkg/db"
	"github.com/mj37yhyy/gowb/pkg/model"
	"github.com/mj37yhyy/gowb/pkg/web/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type HandlerFunc func(context.Context) (model.Response, error)
type Director func(req *http.Request)
type Router struct {
	Path                string
	Method              string
	Handler             HandlerFunc
	OpenFlatTransaction bool
	Director            Director
}

func Bootstrap(ctx context.Context) {
	server := start(ctx)
	_signal()
	_timeout(ctx, server)
}

func start(c context.Context) *http.Server {
	conf := c.Value(constant.ConfigKey).(config.Config)
	routers := c.Value(constant.RoutersKey).([]Router)

	gin.SetMode(conf.Web.RunMode)

	routersInit := doRouter(c, routers)
	readTimeout := time.Minute
	writeTimeout := time.Minute
	endPoint := fmt.Sprintf(":%d", conf.Web.Port)
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
		ctx.Next()
	})
	//r.Use(middleware.RequestLogging())
	r.Use(middleware.Logger())
	r.Use(ginprom.PromMiddleware(nil))
	r.Use(middleware.Tracing())
	return r
}

/**
路由
*/
func router(r *gin.Engine, routers []Router) *gin.Engine {
	baseHandle(r)
	doHandle(r, routers)
	return r
}

/*
基础处理
*/
func baseHandle(r *gin.Engine) {
	// 404 Handler.
	r.NoRoute(func(c *gin.Context) {
		resp := model.Response{}
		resp.SetError(model.ErrorInfo{Code: http.StatusNotFound, Message: "The incorrect API route."})
		c.JSON(http.StatusNotFound, resp)
	})

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf(time.Now().String()))
	})

	r.GET("/metrics", ginprom.PromHandler(promhttp.Handler()))
}

/**
用户函数处理
*/
func doHandle(r *gin.Engine, routers []Router) {
	for _, router := range routers {
		ch := make(chan int)
		go func(_router Router) {
			r.Handle(_router.Method, _router.Path, func(ctx *gin.Context) {
				if router.Director != nil {
					//透传
					proxy := &httputil.ReverseProxy{Director: router.Director}
					proxy.ServeHTTP(ctx.Writer, ctx.Request)
				} else {
					//调用
					c := getContext(ctx)
					getBody(c, ctx)
					getParams(c, ctx)
					getHeader(c, ctx)
					addRequest(c, ctx)
					call(_router, c, ctx)
				}
			})
			ch <- 0
		}(router)
		<-ch
	}
}

/*
调用用户函数
*/
func call(_router Router, c context.Context, ctx *gin.Context) {
	var tx *gorm.DB
	if _router.OpenFlatTransaction {
		tx = db.DB.Begin()
		setContext(ctx, context.WithValue(c, constant.TransactionKey, tx))
	}
	resp, err := _router.Handler(c)

	if err != nil {
		if tx != nil && _router.OpenFlatTransaction {
			tx.Rollback()
		}
		ctx.JSON(resp.Error.Code, resp)
	} else {
		if tx != nil && _router.OpenFlatTransaction {
			tx.Commit()
		}
		ctx.JSON(http.StatusOK, resp)
	}
}

/*
将request放入上下文
*/
func addRequest(c context.Context, ctx *gin.Context) {
	setContext(ctx, context.WithValue(c, constant.RequestKey, ctx.Request))
}

/*
将header放入上下文
*/
func getHeader(c context.Context, ctx *gin.Context) {
	setContext(ctx, context.WithValue(c, constant.HeaderKey, ctx.Request.Header))
}

/*
将参数放入上下文
*/
func getParams(c context.Context, ctx *gin.Context) {
	request := ctx.Request
	var params = make(map[string][]string)

	// url 参数
	for _, param := range ctx.Params {
		params[param.Key] = []string{param.Value}
	}

	// querystring
	for key, val := range request.URL.Query() {
		params[key] = val
	}

	// form data
	for key, val := range request.PostForm {
		params[key] = val
	}

	setContext(ctx, context.WithValue(c, constant.ParamsKey, params))
	return
}

/*
将body放入上下文
*/
func getBody(c context.Context, ctx *gin.Context) {
	body, _ := ioutil.ReadAll(ctx.Request.Body)
	//fmt.Println(body)
	setContext(ctx, context.WithValue(c, constant.BodyKey, body))
}

func getContext(ctx *gin.Context) context.Context {
	return ctx.Value(constant.ContextKey).(context.Context)
}

func setContext(ctx *gin.Context, c context.Context) {
	ctx.Set(constant.ContextKey, c)
}
