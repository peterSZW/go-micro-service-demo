package main

import (
	"fmt"

	"github.com/buaazp/fasthttprouter"
	"github.com/inconshreveable/log15"
	"github.com/valyala/fasthttp"

	"net"

	"flag"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

const serviceName = "userservice"

var AllDone chan bool

// ab -c 10 -n 1000000 http://127.0.0.1:8899/hystrix?percent=24
//service_demo -consul 172.16.6.201:8500 -port 8899

var port int
var consuladdress string

func main() {

	flag.IntVar(&port, "port", 8899, "WEB服务端口号")
	flag.StringVar(&consuladdress, "consul", "127.0.0.1:8500", "Consul IP 端口号")

	flag.Parse()
	log15.Info("Config", "consul", consuladdress, "port", port)

	runtime.GOMAXPROCS(runtime.NumCPU())
	go httpserver()
	AllDone = make(chan bool)
	<-AllDone
}

func echo(c *fasthttp.RequestCtx) {
	qps.Count()
	c.SuccessString("application/json", `{"code":0}`)
}

func health(c *fasthttp.RequestCtx) {
	c.SuccessString("application/json", `{"code":0}`)
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

func httpserver() {

	r := fasthttprouter.New()

	//Qps Monitor Web Service Handle
	qps_init()
	r.GET("/qps", qps_Handle)
	r.GET("/qps_json", qps_json_Handle)

	//Consul Health Check
	r.GET("/health", health)

	//Web Service Handle
	r.GET("/echo", echo)
	r.GET("/echo2", echo2)

	var ln net.Listener
	var err error
	for i := 0; i < 10; i++ {
		ln, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			port = port + 1
		} else {
			break
		}
	}
	checkErr(err, "httpServer", true)

	registerService()
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
			if s != syscall.SIGURG {
				log15.Info("catch exit signal", "get signal", s)
			}

		}
	}
}

func httpget(req *fasthttp.Request, url string) (*fasthttp.Response, error) {
	req.SetRequestURI(url)
	resp := &fasthttp.Response{}
	client := fasthttp.Client{}
	err := client.Do(req, resp)
	return resp, err
}
