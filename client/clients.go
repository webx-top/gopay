package client

import "github.com/webx-top/gopay/common"

var clients = map[string]func() common.PayClient{}

func Register(name string, engine func() common.PayClient) {
	clients[name] = engine
}

func Get(name string) func() common.PayClient {
	engine, _ := clients[name]
	return engine
}
