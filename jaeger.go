package main

import (
	"net"

	"context"
	"fmt"
	"github.com/d7561985/opentracefasthttp"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/stretchr/testify/assert"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"io"
	"time"
)

//https://blog.csdn.net/liyunlong41/article/details/87932953

func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:           false,
			LocalAgentHostPort: "172.16.1.216:6831", //http://172.16.1.216/
		},
	}
	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}
func foo3(req string, ctx context.Context) (reply string) {
	//1.创建子span
	span, _ := opentracing.StartSpanFromContext(ctx, "span_foo3")
	defer func() {
		//4.接口调用完，在tag中设置request和reply
		span.SetTag("request", req)
		span.SetTag("reply", reply)
		span.Finish()
	}()

	println(req)
	//2.模拟处理耗时
	time.Sleep(time.Second / 2)
	//3.返回reply
	reply = "foo3Reply"
	return
}

//跟foo3一样逻辑
func foo4(req string, ctx context.Context) (reply string) {
	span, _ := opentracing.StartSpanFromContext(ctx, "span_foo4")
	defer func() {
		span.SetTag("request", req)
		span.SetTag("reply", reply)
		span.Finish()
	}()

	println(req)
	time.Sleep(time.Second / 2)
	reply = "foo4Reply"
	return
}

func main2() {
	tracer, closer := initJaeger("jaeger-demo")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer) //StartspanFromContext创建新span时会用到

	span := tracer.StartSpan("span_root")
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	r1 := foo3("Hello foo3", ctx)
	r2 := foo4("Hello foo4", ctx)
	fmt.Println(r1, r2)
	span.Finish()
}
func main1() {
	main2()

	var t assert.TestingT

	//tracer, closer := jaeger.NewTracer("fasthttp-carrier-tester", jaeger.NewConstSampler(true), jaeger.NewNullReporter())

	tracer, closer := initJaeger("fasthttp-carrier-tester")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	ok := false
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close() //nolint

	srv := fasthttp.Server{Handler: func(ctx *fasthttp.RequestCtx) {
		carrier := opentracefasthttp.New(&ctx.Request.Header)
		clientContext, err := tracer.Extract(opentracing.HTTPHeaders, carrier)

		fmt.Println(ctx.Request.Header.String())
		assert.NoError(t, err)
		span := tracer.StartSpan("HTTP "+string(ctx.Method())+" "+ctx.Request.URI().String(), ext.RPCServerOption(clientContext))

		fmt.Println("3", span)

		assert.NotNil(t, span)
		span.LogFields(log.String("server", "request ok"))
		fmt.Println("4", span)
		defer span.Finish()

		ok = true
	}}
	go srv.Serve(ln) //nolint

	span := opentracing.GlobalTracer().StartSpan("client-request")
	defer span.Finish()

	fmt.Println("1", span)

	span.SetTag("test", "test")
	span.LogFields(log.String("test", "test"))

	fmt.Println("2", span)

	req := fasthttp.AcquireRequest()
	req.Header.SetMethod(fasthttp.MethodGet)
	req.SetRequestURI("http://example.com")

	carrier := opentracefasthttp.New(&req.Header)
	err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	assert.NoError(t, err)

	client := fasthttp.Client{Dial: func(addr string) (net.Conn, error) {
		return ln.Dial()
	}}
	err = client.Do(req, nil)
	assert.NoError(t, err)

	assert.True(t, ok)
}
