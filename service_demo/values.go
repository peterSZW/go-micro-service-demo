// main
package main

import (
	"container/list"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"time"
)

const FeatureCNT = 100
const UserCNT = 5000000
const ResultCNT = 100

type peopledata struct {
	Ans [FeatureCNT]int
}

type TSearchres struct {
	ids    [ResultCNT]int
	dis    [ResultCNT]int
	mindis int
	cnt    int
}

func (this *TSearchres) add(id int, dis int) {

	tobeCompare := 0

	if this.cnt == ResultCNT {
		if this.mindis > dis {
			log.Println("all bigger ", this.mindis, dis)
			return
		}
		tobeCompare = this.mindis
		log.Println("conver ", this.mindis, dis)

	}

	for i := 0; i < ResultCNT; i++ {

		if this.dis[i] == tobeCompare {
			this.ids[i] = id
			this.dis[i] = dis
			if this.mindis == 0 { //第一次加进来
				this.mindis = dis
			}
			if this.mindis > dis { //非0
				this.mindis = dis
			}
			if this.cnt < ResultCNT {
				this.cnt++
			}
			log.Println("add ", i, id, dis, this.mindis, this.cnt)
			return

		}
	}
}

func (this *TSearchres) list() {
	log.Println("==========")
	for i := 0; i < ResultCNT; i++ {
		if this.dis[i] > 0 {
			fmt.Println(this.ids[i], this.dis[i])
		}
	}
	log.Println("==========")
}
func (this *TSearchres) clear() {
	this.mindis = 0
	this.cnt = 0
	for i := 0; i < ResultCNT; i++ {
		this.ids[i] = 0
		this.dis[i] = 0
	}
}
func calcDis(a peopledata, b peopledata) int {
	cnt := 0
	for i := 0; i < FeatureCNT; i++ {
		if a.Ans[i] == b.Ans[i] {
			cnt++
		}
	}
	return cnt

}

func RandomDis(a *peopledata) {

	for i := 1; i < FeatureCNT; i++ {
		a.Ans[i] = rand.Intn(2)
	}
}

func usearry() {

	var a [UserCNT]peopledata
	{
		start := time.Now()
		for i := 0; i < UserCNT; i++ {

			a[i].Ans[0] = i
			RandomDis(&a[i])

			if i%1000000 == 0 {
				fmt.Println(i)
			}
		}
		log.Println("OK", time.Since(start))
	}
	{

		var user peopledata
		RandomDis(&user)

		start := time.Now()

		var searchresult TSearchres

		for i := 0; i < UserCNT; i++ {

			dis := calcDis(user, a[i])
			if dis > *rate {
				log.Println(dis, "=", a[i], user)
				searchresult.add(a[i].Ans[0], dis)
			}

			if i%1000000 == 0 {
				fmt.Println(i)
			}
		}

		log.Println("OK", time.Since(start))
		searchresult.list()
	}
}
func timmer(f func(), title string) {
	fmt.Println(title, "started...")
	start := time.Now()
	f()
	log.Println(title, "finished.", time.Since(start))

}

var rate *int = flag.Int("rate", 70, "same rate")

func valuemain() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())

	timmer(usearry, "usearry")
}
func uselist() {
	peoples := list.New()
	log.Println("Hello World!111")
	for i := 0; i < 100000000; i++ {

		people := new(peopledata)
		people.Ans[1] = i
		peoples.PushBack(people)
		if i%1000000 == 0 {
			fmt.Println(i)
		}
	}
	fmt.Println("Hello World!")
	for l := peoples.Front(); l != nil; l = l.Next() {
		if (l.Value).(*peopledata).Ans[1] == 90000000 {
			fmt.Println((l.Value).(*peopledata))
		}

	}
	log.Println("Hello World!")
}
