# 微服务的客户端

程序启动做以下工作


1. 服务提供基础的服务  echo

这个服务端没有任何限速和熔断限制，只有 HealthCheck，我们可以在 HealthCheck 里检查我们的 CPU 内存，返回消息到服务器。

## 