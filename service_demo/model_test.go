package main

import (
	"fmt"
	"testing"
)

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

func Test_Person(t *testing.T) {
	var person = TPerson{DocId: 12}

	t.Log(person.AsJsonString())

}
