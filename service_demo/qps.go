package main

import (
	"github.com/inconshreveable/log15"
	"github.com/valyala/fasthttp"
	mqps "github.com/zengming00/go-qps"
	"math/rand"
	"time"
)

var qps *mqps.QP

func qps_Handle(c *fasthttp.RequestCtx) {
	s, err := qps.Show()
	if err != nil {
		log15.Error("qps show error", "err", err)
	}
	c.SuccessString("text/html; charset=utf-8", s)
}

func qps_json_Handle(c *fasthttp.RequestCtx) {
	bts, err := qps.GetJson()
	if err != nil {
		log15.Error("qps GetJson error", "err", err)
	}
	c.SuccessString("application/json", string(bts))
}

func qps_init() {
	rand.Seed(time.Now().UnixNano())
	// Statistics every second, a total of 3600 data
	qps = mqps.NewQP(time.Second, 3600)
}
