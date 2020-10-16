package robinwithweight

import (
	"fmt"
	"sync"
)

type Service struct {
	ServiceID string //url连接
	Address   string
	Port      int
	Weight    int64 //权重
	CallTimes int64 //调用次数统计
}
type RobinwithWeight struct {
	services []*Service
	weitht   int64
	index    int64
	lock     sync.Mutex
}

func (c *RobinwithWeight) findService(serviceID string) int {

	for k, v := range c.services {
		if v.ServiceID == serviceID {
			return k
		}
	}
	return -1
}

func (c *RobinwithWeight) Delservice(serviceID string) int {
	c.lock.Lock()
	defer c.lock.Unlock()

	for k, v := range c.services {
		if v.ServiceID == serviceID {
			c.services = append(c.services[:k], c.services[k+1:]...)
			return k
		}
	}
	return -1

}

func (c *RobinwithWeight) Addservice(serviceID string, address string, port int, weight int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.services) == 0 && c.index == 0 {
		c.index = -1 //for init
	}
	i := c.findService(serviceID)
	if i >= 0 {
		c.services[i].Port = port
		c.services[i].Address = address
		c.services[i].Weight = weight
		return
	} else {
		c.services = append(c.services, &Service{ServiceID: serviceID, Address: address, Port: port, Weight: weight, CallTimes: 0})
		return
	}
}

//计算两个数的最大公约数
func (c *RobinwithWeight) gcd(a, b int64) int64 {
	if b == 0 {
		return a
	}
	return c.gcd(b, a%b)
}

//计算多个数的最大公约数
func (c *RobinwithWeight) getGCD() int64 {
	var weights []int64

	for _, url := range c.services {
		weights = append(weights, url.Weight)
	}

	if len(weights) == 0 { //可能没有url
		return 1
	}

	g := weights[0]
	for i := 1; i < len(weights)-1; i++ {
		oldGcd := g
		g = c.gcd(oldGcd, weights[i])
	}
	return g
}

//获取当前轮次的url
func (c *RobinwithWeight) GetUrl() *Service {

	if len(c.services) == 0 {
		return nil
	}

	gcd := c.getGCD()
	for {
		c.index = (c.index + 1) % int64(len(c.services))
		if c.index == 0 {
			c.weitht = c.weitht - gcd
			if c.weitht <= 0 {
				c.weitht = c.getMaxWeight()
				if c.weitht == 0 {
					return &Service{}
				}
			}
		}

		if c.services[c.index].Weight >= c.weitht {
			return c.services[c.index]
		}
	}
}

//获取最大权重
func (c *RobinwithWeight) getMaxWeight() int64 {
	var max int64 = 0
	for _, url := range c.services {
		if url.Weight >= int64(max) {
			max = url.Weight
		}
	}

	return max
}

//测试函数
func test_main() {
	//模拟1000次调用，统计每个url的调用次数
	var robin RobinwithWeight
	robin.Addservice("127.0.0.1-8080", "127.0.0.1", 8080, 2)
	robin.Addservice("127.0.0.1-8081", "127.0.0.1", 8081, 4)
	robin.Addservice("127.0.0.1-8090", "127.0.0.1", 8090, 4)

	for i := 0; i < 1000; i++ {
		url := robin.GetUrl()
		//fmt.Println("call: ", url.ServiceID)
		if url.CallTimes != 0 {
			url.CallTimes++
		} else {
			url.CallTimes = 1
		}
	}

	fmt.Printf("result : \n")
	for i, url := range robin.services {
		fmt.Println(i, url.ServiceID, url.Weight, url.CallTimes)
	}
	robin.Delservice("127.0.0.1-8090")

	for i := 0; i < 1000; i++ {
		url := robin.GetUrl()
		//fmt.Println("call: ", url.ServiceID)
		if url.CallTimes != 0 {
			url.CallTimes++
		} else {
			url.CallTimes = 1
		}
	}

	fmt.Printf("result : \n")
	for i, url := range robin.services {
		fmt.Println(i, url.ServiceID, url.Weight, url.CallTimes)
	}
}
