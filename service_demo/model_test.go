package main

import (
	"fmt"
	"testing"
)

var person = Person{DocId: 123, Position: "搜索工程师", Company: "百度", City: "北京", SchoolLevel: 2, Vip: false, Chat: true, Active: 1, WorkAge: 3}

func Test_Jsoniter(t *testing.T) {
	type ColorGroup struct {
		ID     int
		Name   string
		Colors []string
	}
	group := ColorGroup{
		ID:     1,
		Name:   "Reds",
		Colors: []string{"Crimson", "Red", "Ruby", "Maroon"},
	}

	b, err := json_iterator.Marshal(group)
	t.Log(string(b))
	if err != nil {
		fmt.Println("error:", err)
	}

}
