package main

import (
	"github.com/inconshreveable/log15"
	jsoniter "github.com/json-iterator/go"
)

type TPerson struct {
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

func (this *TPerson) add(id int, dis int) {

}

func (this *TPerson) AsJsonString() string {

	b, err := json_iterator.Marshal(this)
	if err != nil {
		log15.Error("JSON Marshal", "errr", err, "person", this)
		return ""
	} else {
		return string(b)
	}
}
