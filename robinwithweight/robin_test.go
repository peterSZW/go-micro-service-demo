package robinwithweight

import (
	"fmt"
	"testing"
)

//测试函数
func TestMain(t *testing.T) {
	//模拟1000次调用，统计每个url的调用次数
	var robin RobinwithWeight
	robin.Addservice("127.0.0.1-8080", "127.0.0.1", 8080, 2)
	robin.Addservice("127.0.0.1-8081", "127.0.0.1", 8081, 4)
	robin.Addservice("127.0.0.1-8090", "127.0.0.1", 8090, 4)

	for i := 0; i < 1000; i++ {
		url := robin.GetUrl()
		//fmt.Println("call: ", url.Url)
		if url.CallTimes != 0 {
			url.CallTimes++
		} else {
			url.CallTimes = 1
		}
	}

	fmt.Printf("result : \n")
	for i, url := range robin.services {
		fmt.Println(i, url.ServiceID, url.Weight, url.CallTimes)
		t.Log(i, url.ServiceID, url.Weight, url.CallTimes)
	}
	robin.Delservice("127.0.0.1-8090")

	for i := 0; i < 1000; i++ {
		url := robin.GetUrl()
		//fmt.Println("call: ", url.Url)
		if url.CallTimes != 0 {
			url.CallTimes++
		} else {
			url.CallTimes = 1
		}
	}

	fmt.Printf("result : \n")
	for i, url := range robin.services {
		fmt.Println(i, url.ServiceID, url.Weight, url.CallTimes)
		t.Log(i, url.ServiceID, url.Weight, url.CallTimes)
	}
}
