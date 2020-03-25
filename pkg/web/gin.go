package web

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chenjiandongx/ginprom"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	"github.com/mj37yhyy/gowb/pkg/config"
	"github.com/mj37yhyy/gowb/pkg/constant"
	"github.com/mj37yhyy/gowb/pkg/db"
	"github.com/mj37yhyy/gowb/pkg/model"
	"github.com/mj37yhyy/gowb/pkg/web/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

type HttpStatus int
type HandlerFunc func(context.Context) (model.Response, HttpStatus)
type Director func(req *http.Request) func(req *http.Request)
type Router struct {
	Path                string
	Method              string
	Handler             HandlerFunc
	OpenFlatTransaction bool
	ReverseProxy        bool
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

	_config := c.Value(constant.ConfigKey).(config.Config)

	r.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: _config.Web.LogSkipPath,
	}))
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
		resp.SetError(model.ErrorInfo{
			Code:    "NotFound",
			Message: "The incorrect API route."})
		c.JSON(http.StatusNotFound, resp)
	})

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, fmt.Sprintf(time.Now().String()))
	})

	r.GET("/metrics", ginprom.PromHandler(promhttp.Handler()))
	r.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

/**
用户函数处理
*/
func doHandle(r *gin.Engine, routers []Router) {
	for _, router := range routers {
		ch := make(chan int)
		go func(_router Router) {
			r.Handle(_router.Method, _router.Path, func(ctx *gin.Context) {
				if _router.ReverseProxy {
					//透传
					proxy := &httputil.ReverseProxy{Director: router.Director(ctx.Request)}
					proxy.ServeHTTP(ctx.Writer, ctx.Request)
				} else {
					//调用
					addBody(ctx)
					addParams(ctx)
					addHeader(ctx)
					addRequest(ctx)
					addShouldBind(ctx)
					addBind(ctx)
					call(_router, ctx)
				}
			})
			ch <- 0
		}(router)
		<-ch
	}
}

func addShouldBind(ctx *gin.Context) {
	setContext(ctx, context.WithValue(getContext(ctx), constant.ShouldBindKey, func(obj interface{}) error {
		return ctx.ShouldBind(obj)
	}))
	setContext(ctx, context.WithValue(getContext(ctx), constant.ShouldBindWithKey,
		func(obj interface{}, bt constant.BindingType) error {
			return ctx.ShouldBindWith(obj, getBinding(bt))
		}))
}

func addBind(ctx *gin.Context) {
	setContext(ctx, context.WithValue(getContext(ctx), constant.BindKey, func(obj interface{}) error {
		return ctx.Bind(obj)
	}))
	setContext(ctx, context.WithValue(getContext(ctx), constant.BindWithKey,
		func(obj interface{}, bt constant.BindingType) error {
			return ctx.MustBindWith(obj, getBinding(bt))
		}))
}

func getBinding(bt constant.BindingType) binding.Binding {
	switch bt {
	case constant.BindingForm:
		return binding.Form
	case constant.BindingFormPost:
		return binding.FormPost
	case constant.BindingFormMultipart:
		return binding.FormMultipart
	case constant.BindingQuery:
		return binding.Query
	case constant.BindingHeader:
		return binding.Header
	case constant.BindingJson:
		return binding.JSON
	case constant.BindingYaml:
		return binding.YAML
	case constant.BindingXml:
		return binding.XML
	case constant.BindingMsgPack:
		return binding.MsgPack
	case constant.BindingProtoBuf:
		return binding.ProtoBuf
	default:
		return binding.Form
	}
}

/*
调用用户函数
*/
func call(_router Router, ctx *gin.Context) {
	var tx *gorm.DB
	if _router.OpenFlatTransaction {
		tx = db.DB.Begin()
		setContext(ctx, context.WithValue(getContext(ctx), constant.TransactionKey, tx))
	}
	resp, hs := _router.Handler(getContext(ctx))

	if hs >= 400 {
		if tx != nil && _router.OpenFlatTransaction {
			tx.Rollback()
		}
		if unsafe.Sizeof(resp) > 0 {
			ctx.JSON(int(hs), resp)
		}
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
func addRequest(ctx *gin.Context) {
	setContext(ctx, context.WithValue(getContext(ctx), constant.RequestKey, ctx.Request))
}

/*
将header放入上下文
*/
func addHeader(ctx *gin.Context) {
	setContext(ctx, context.WithValue(getContext(ctx), constant.HeaderKey, ctx.Request.Header))
}

/*
将参数放入上下文
*/
func addParams(ctx *gin.Context) {
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

	setContext(ctx, context.WithValue(getContext(ctx), constant.ParamsKey, params))
	return
}

/*
将body放入上下文
*/
func addBody(ctx *gin.Context) {
	body, _ := ioutil.ReadAll(ctx.Request.Body)
	ctx.Request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	setContext(ctx, context.WithValue(getContext(ctx), constant.BodyKey, body))
}

func getContext(ctx *gin.Context) context.Context {
	return ctx.Value(constant.ContextKey).(context.Context)
}

func setContext(ctx *gin.Context, c context.Context) {
	ctx.Set(constant.ContextKey, c)
}
