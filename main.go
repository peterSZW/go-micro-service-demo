package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/afex/hystrix-go/hystrix"
	"github.com/buaazp/fasthttprouter"
	"github.com/d7561985/opentracefasthttp"
	"github.com/inconshreveable/log15"
	"github.com/valyala/fasthttp"
	"go.uber.org/ratelimit"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	//	"github.com/opentracing/opentracing-go/log"
	// "github.com/valyala/fasthttp"
	// "github.com/valyala/fasthttp/fasthttputil"
	// "io"
)

const serviceName = "gateway"

//ab -c 10 -n 10000 http://127.0.0.1:8899/app/userservice/echo
//ab -c 10 -n 10000 http://127.0.0.1:8899/app/gateway/echo
//consul.exe agent -server -bootstrap -data-dir G:\consul\data -bind="127.0.0.1" -client="0.0.0.0" -node=abc -datacenter=goMicroService -ui

func httpget(req *fasthttp.Request, url string) (*fasthttp.Response, error) {
	req.SetRequestURI(url)
	resp := &fasthttp.Response{}
	client := fasthttp.Client{}
	err := client.Do(req, resp)
	return resp, err
}

func echo2(ctx *fasthttp.RequestCtx) {
	url := "http://127.0.0.1:8899/app/gateway/echo"
	resp, err := httpget(&ctx.Request, url)

	qps.Count()

	if err != nil {
		log15.Error("fasthttp Call error", "err", err)
	} else {
		ctx.SuccessString("application/json", string(resp.Body()))

	}

}

func echo(c *fasthttp.RequestCtx) {

	qps.Count()
	c.SuccessString("application/json", `{"code":0}`)
}

func health(c *fasthttp.RequestCtx) {

	UpdateAgentServices()
	c.SuccessString("application/json", `{"code":0}`)
}

func limit(c *fasthttp.RequestCtx) {
	ratelimitClient.Take()
	qps.Count()

	c.SuccessString("application/json", `{"code":0}`)
}

//ab -c 6 -n 10000000  http://127.0.0.1:8899/hystrix?percent=20

func hystrixtest(c *fasthttp.RequestCtx) {
	qps.Count()

	percent := c.QueryArgs().GetUintOrZero("percent")

	hystrix.Do("hystrixtest", func() error {
		n := rand.Intn(100)

		if false {
			time.Sleep(1 * time.Second)
		}

		if n < percent {
			c.SuccessString("application/json", fmt.Sprintf(`{"MSG":"ERR","rand":%d,"percent":%d,	 }`, n, percent))
			return errors.New("hystrixtest ERR" + fmt.Sprintf(`{"MSG":"ERR","rand":%d,"percent":%d,	 }`, n, percent))
		} else {
			c.SuccessString("application/json", fmt.Sprintf(`{"MSG":"OK","rand":%d,"percent":%d,	 }`, n, percent))
			return nil
		}

	}, func(err error) error {

		c.SuccessString("application/json", `{"hystrixtest back":0}`+err.Error())
		return nil
	})

}

func index(c *fasthttp.RequestCtx) {
	c.SuccessString("text/html;charset=utf-8", "access <a href=/app/service/method>/app/service/method</a> to call your service method.")
}

func deregister(c *fasthttp.RequestCtx) {
	deregisterService()
	c.SuccessString("application/json", `{"code":0}`)
}

func register(c *fasthttp.RequestCtx) {
	registerService()

	c.SuccessString("application/json", `{"code":0}`)
}
func listAllService() {
	AgentService := GetAgentServices()
	for k, v := range AgentService {
		s, _ := json.Marshal(v)
		fmt.Printf("%s %v\n", k, string(s))
	}
}

func callOtherMicroService(ctx *fasthttp.RequestCtx) {

	servicename, methodname := ctx.UserValue("service").(string), ctx.UserValue("method").(string)
	//fmt.Println(ctx.Request.Header.String())

	var span opentracing.Span
	{
		carrier := opentracefasthttp.New(&ctx.Request.Header)
		clientContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)

		if err != nil {
			name := servicename + "-" + methodname

			span = opentracing.GlobalTracer().StartSpan(name)
			//log15.Info("StartNew Span", "span", span, "name", name, "err", err)
			//span.LogFields(log.String("server", "request ok"))

		} else {
			name := servicename + "-" + methodname //name = "HTTP "+string(ctx.Method())+" "+ctx.Request.URI().String()
			span = opentracing.GlobalTracer().StartSpan(name, ext.RPCServerOption(clientContext))
			//log15.Info("Continue  Span", "span", span, "name", name, "err", err)
			//span.LogFields(log.String("server", "request ok"))

		}
	}
	defer span.Finish()
	service := GetServiceFromMapWithRoundRobin(servicename)
	if service != nil {
		url := fmt.Sprintf("http://%s:%d/%s", service.Address, service.Port, methodname)
		//fmt.Println(url)

		WGCallService.Add(1)
		defer WGCallService.Done()

		carrier := opentracefasthttp.New(&ctx.Request.Header)
		err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
		rsp, err := httpget(&ctx.Request, url)

		if err == nil {
			data := rsp.Body()
			ctx.SuccessString("application/json", string(data))
			return
		} else {

			ctx.SuccessString("application/json", err.Error())

			return
		}
	}

	ctx.SuccessString("application/json", `{"err":"service not found"}`)

}

var ratelimitClient ratelimit.Limiter

var AllDone chan bool

var port int
var consuladdress string

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	{
		//跟踪 opentracing
		tracer, closer := initJaeger("micro-service")
		defer closer.Close()
		opentracing.SetGlobalTracer(tracer)
	}
	{
		//命令行参数
		flag.IntVar(&port, "port", 8899, "WEB服务端口号")
		flag.StringVar(&consuladdress, "consul", "127.0.0.1:8500", "Consul IP 端口号")
		flag.Parse()
		log15.Info("Config", "consul", consuladdress, "port", port)
	}
	{
		//限速
		ratelimitClient = ratelimit.New(500)

		//熔断
		hystrix.ConfigureCommand("hystrixtest", hystrix.CommandConfig{
			Timeout:               1000,
			MaxConcurrentRequests: 100,
			ErrorPercentThreshold: 25,
		})
		// Timeout: 执行command的超时时间。默认时间是1000毫秒
		// MaxConcurrentRequests：command的最大并发量 默认值是10
		// SleepWindow：当熔断器被打开后，SleepWindow的时间就是控制过多久后去尝试服务是否可用了。默认值是5000毫秒
		// RequestVolumeThreshold： 一个统计窗口10秒内请求数量。达到这个请求数量后才去判断是否要开启熔断。默认值是20
		// ErrorPercentThreshold：错误百分比，请求数量大于等于RequestVolumeThreshold并且错误率到达这个百分比后就会启动熔断 默认值是50

		//hystrix 状态服务
		hystrixStreamHandler := hystrix.NewStreamHandler()
		hystrixStreamHandler.Start()
		go http.ListenAndServe(net.JoinHostPort("", "8182"), hystrixStreamHandler)

		// ab -c 10 -n 1000000 http://127.0.0.1:8899/hystrix?percent=24
		// TO MONITOR	[root@hotmall_dev:~]# docker run -d -p 9090:9002 --name hystrix-dashboard mlabouardy/hystrix-dashboard:latest
		// http://dockerhost:9090/hystrix
		// http://127.0.0.1:8182
	}

	go updateServiceEvery5Second()
	go httpserver()

	AllDone = make(chan bool)
	<-AllDone

}

func httpserver() {

	r := fasthttprouter.New()
	qps_init()
	r.GET("/qps", qps_Handle)
	r.GET("/qps_json", qps_json_Handle)

	r.GET("/app/:service/:method", callOtherMicroService)
	r.POST("/app/:service/:method", callOtherMicroService)
	// r.GET("/register", register)
	// r.GET("/deregister", deregister)
	r.GET("/echo", echo)
	r.GET("/echo2", echo2)
	r.GET("/health", health)
	r.GET("/limit", limit)
	r.GET("/hystrix", hystrixtest)
	r.GET("/", index)
	port = 8899

	var ln net.Listener
	var err error
	for i := 0; i < 3; i++ {
		ln, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			port = port + 1
		} else {
			break
		}
	}
	checkErr(err, "httpServer", true)

	registerService()
	UpdateAgentServices()
	UpdateAllHealthServices()
	GenServiceMap() // Gen robin map

	go CatchExitSignal(ln)

	log15.Info("http start listen", "port", port)
	if err := fasthttp.Serve(ln, r.Handler); err != nil {
		checkErr(err, "httpServer", true)
	}
}

func checkErr(err error, errMessage string, isQuit bool) {
	if err != nil {
		log15.Error(errMessage, "error", err)
		if isQuit {
			os.Exit(1)
		}
	}
}

var WGCallService sync.WaitGroup

func CatchExitSignal(serverlisten net.Listener) {
	c := make(chan os.Signal)
	signal.Notify(c)
	for {
		s := <-c

		if s == syscall.SIGINT || s == syscall.SIGQUIT || s == syscall.SIGTERM {
			log15.Info("catch exit signal", "get signal", s)

			if serverlisten != nil {
				// 关闭监听端口
				if err := serverlisten.Close(); err != nil {
					log15.Error("serverlisten shutdown error", "err", err)
				}

				// 等待所有处理结束，最长等待10秒

				flagChan := make(chan struct{}, 1)
				go func() {
					WGCallService.Wait()
					flagChan <- struct{}{}
				}()
				select {
				case <-flagChan:
					log15.Info("all request is finished")
				case <-time.After(10 * time.Second):
					log15.Error(("Close after 10 seconds timeout"))
				}

			}
			deregisterService()

			AllDone <- true
		} else {
			log15.Info("catch exit signal", "get signal", s)
		}
	}
}
