# go-micro-service-demo

This is a demo of go micro service. using consul, opentracking, jaeger, hystrix, fasthttp, uber ratelimit, log15.

这是一个 go 的 MicroService 简单的样例，使用的组件如下：

* 服务发现： consul
* 服务跟踪： opentracking, jaeger 
* 限流熔断： hystrix, uber ratelimit
* 网站服务： fasthttp
* 日志组件： log15

考虑到简化样例，当中没有用到任何的 RPC

# Architect 架构

* 当前架构

![avatar](/image/ms1.jpg)

* 可修改为 Sidecar 模式

![avatar](/image/ms2.jpg)

* 也可修改为整合模式

![avatar](/image/ms3.jpg)

# Run 运行 (windows)

* consul

```
consul.exe agent -server -bootstrap -data-dir d:\consul\data -bind="127.0.0.1" -client="0.0.0.0" -node=abc -datacenter=goMicroService -ui
```

* gateway

```
go build
go-micro-service-demo.exe
```

* service

```
cd service_demo
go build
service_demo.exe
```

* ab test 压力测试, 服务跟踪

```
ab -c 10 -n 10000 http://127.0.0.1:8899/echo
ab -c 10 -n 10000 http://127.0.0.1:8899/app/userservice/echo
ab -c 10 -n 10000 http://127.0.0.1:8899/app/userservice/echo2
```

* 限流熔断
```
ab -c 10 -n 100000   http://127.0.0.1:8899/hystrix?percent=20
ab -c 10 -n 100000   http://127.0.0.1:8899/limit
```

# hystrix 面板
```
docker run -d -p 9090:9002 --name hystrix-dashboard mlabouardy/hystrix-dashboard:latest
浏览器访问 http://dockerhost:9090/hystrix
输入 http://127.0.0.1:8182  (gateway IP)
```

# qps 面板
```
浏览器访问 http://dockerhost:9090/qps
```

# jaeger 面板

```
#!/usr/bin/env bash
docker run -d --name jaeger \
    -e COLLECTOR_ZIPKIN_HTTP_PORT=9411 \
    -p 5775:5775/udp \
    -p 6831:6831/udp \
    -p 6832:6832/udp \
    -p 5778:5778 \
    -p 16686:16686 \
    -p 14268:14268 \
    -p 9411:9411 \
    --restart=always \
    jaegertracing/all-in-one:1.15

```

```
修改 jaeger.go 里的 LocalAgentHostPort，指向 jaeger 

Reporter: &config.ReporterConfig{
			LogSpans:           false,
			LocalAgentHostPort: "172.16.1.216:6831",  
		},

```
```
浏览器访问 http://dockerhost:16686/
```

