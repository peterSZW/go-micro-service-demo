package main

import (
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/inconshreveable/log15"
	"sync"
	"time"
)

//======================
var client *consulapi.Client

func Get_consulapi_Client() *consulapi.Client {
	if client == nil {
		var err error
		config := consulapi.DefaultConfig()
		config.Address = consuladdress            //"172.16.6.201:8500"
		client, err = consulapi.NewClient(config) //创建客户端
		if err != nil {
			log15.Crit("Get_consulapi_Client", "err", err)
		}
	}
	return client
}

//http://127.0.0.1:8500/v1/health/state/passing

func RegService(serviceName string) error {

	ip := getHostIP()
	reg := consulapi.AgentServiceRegistration{}
	reg.Name = serviceName //注册service的名字
	reg.Address = ip       //注册service的ip
	reg.Port = port        //注册service的端口
	reg.Tags = []string{"primary"}
	reg.Weights = &consulapi.AgentWeights{}
	reg.Weights.Passing = 2

	check := consulapi.AgentServiceCheck{}                    //创建consul的检查器
	check.Interval = "1s"                                     //设置consul心跳检查时间间隔
	check.HTTP = fmt.Sprintf("http://%s:%d/health", ip, port) //设置检查使用的url

	reg.Check = &check
	reg.ID = fmt.Sprintf("%s-%s-%d", serviceName, ip, port)

	err := Get_consulapi_Client().Agent().ServiceRegister(&reg)
	if err != nil {
		log15.Crit("Get_consulapi_Client", "err", err)
	}
	return err
}

//=================

var agentServices map[string]*consulapi.AgentService

var healthchecks []*consulapi.HealthCheck
var allhealthchecks []*consulapi.HealthCheck

var agentServiceslock sync.Mutex

func updateServiceEvery5Second() {
	for {
		UpdateAgentServices()
		//UpdateHealthServices()
		UpdateAllHealthServices()
		//GenServiceMap() // Gen robin map

		time.Sleep(2 * time.Second)

		//log15.Info("update service list every 5 second")
	}
}

// 请求 Consul 得到服务列表，全局 agentServices 指向新服务列表
func UpdateAgentServices() {
	agentServiceslock.Lock()
	defer agentServiceslock.Unlock()
	newAgentService, err := Get_consulapi_Client().Agent().Services()
	if err == nil {
		agentServices = newAgentService
	}
}

func UpdateHealthServices() {
	agentServiceslock.Lock()
	defer agentServiceslock.Unlock()
	newhealthchecks, _, err := Get_consulapi_Client().Health().State("passing", nil)
	if err == nil {
		healthchecks = newhealthchecks
	}
}

func UpdateAllHealthServices() {
	agentServiceslock.Lock()
	defer agentServiceslock.Unlock()
	newallhealthchecks, _, err := Get_consulapi_Client().Health().State("any", nil)
	if err == nil {
		allhealthchecks = newallhealthchecks
	}

}

// 得到 agentServices 服务列表
func GetAgentServices() map[string]*consulapi.AgentService {
	agentServiceslock.Lock()
	defer agentServiceslock.Unlock()
	return agentServices
}

func deregisterService() error {
	var err error
	ip := getHostIP()
	serviceID := fmt.Sprintf("%s-%s-%d", serviceName, ip, port)

	if err = Get_consulapi_Client().Agent().ServiceDeregister(serviceID); err != nil {
		log15.Error("deregister error", "serviceID", serviceID, "err", err)
	} else {
		log15.Info("deregister success", "serviceID", serviceID)

	}
	return err

}
func registerService() error {
	var err error
	if err = RegService(serviceName); err != nil {
		log15.Error("register error", "serviceName", serviceName, "error", err)
	} else {
		log15.Info("register success", "serviceName", serviceName)
	}
	return err

}

//======================

func KVWrite(key, value string) {

	p := &consulapi.KVPair{Key: key, Flags: 42, Value: []byte(value)}
	if _, err := Get_consulapi_Client().KV().Put(p, nil); err != nil {
		fmt.Println(err)
	}
	fmt.Println(p)

}

func KVRead(key string) {
	pair, _, err := Get_consulapi_Client().KV().Get(key, nil)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pair)
}

//=================
