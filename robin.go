package main

import (
	"fmt"
	"github.com/inconshreveable/log15"
	"go-micro-service-demo/robinwithweight"
)

var serviceMap map[string]*robinwithweight.RobinwithWeight

func init() {
	serviceMap = make(map[string]*robinwithweight.RobinwithWeight)
}
func GetServiceFromMapWithRoundRobin(serviceName string) *robinwithweight.Service {

	a := serviceMap[serviceName]
	if a != nil {
		return a.GetUrl()
	} else {

	}
	return nil

}

func GenServiceMap() {

	serviceMap = make(map[string]*robinwithweight.RobinwithWeight) //renew a map ,because if a service dereg,all health wont have that service
	fmt.Println("+----------------- GenServiceMap ------------------------+")

	for _, v := range allhealthchecks {

		if v.ServiceID == "" {
			continue
		}

		fmt.Println("| ", v.ServiceName, v.ServiceID, v.Status)

		a := serviceMap[v.ServiceName]
		if a == nil {
			a = &robinwithweight.RobinwithWeight{}
			serviceMap[v.ServiceName] = a
		}

		if v.Status == "passing" {
			x := agentServices[v.ServiceID]
			if x == nil {
				log15.Error("agentServices get error", "ServiceID", v.ServiceID)

			} else {
				a.Addservice(v.ServiceID, x.Address, x.Port, 1)
			}
		} else {
			a.Delservice(v.ServiceID)
		}

	}
	fmt.Println("+--------------------------------------------------------+")

}
