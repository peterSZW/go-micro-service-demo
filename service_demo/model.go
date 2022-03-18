package main

import jsoniter "github.com/json-iterator/go"

type Person struct {
	DocId       uint32
	Position    string
	Company     string
	City        string
	SchoolLevel int32
	Vip         bool
	Chat        bool
	Active      int32
	WorkAge     int32
}

var json_iterator = jsoniter.ConfigCompatibleWithStandardLibrary

func test() {

}
