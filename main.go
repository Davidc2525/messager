package main

import (
	"github.com/Davidc2525/messager/core"
	"github.com/Davidc2525/messager/log"

	//"github.com/coreos/etcd"
)

var (
	log = mlog.New()
)

func main() {
	core.Start()
}
