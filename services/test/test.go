package test

import (
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/services/test/src"
	//"sync"
)

type Test struct {

	//mux sync.Mutex
}

var (
	log = mlog.New()
)

func (this *Test) Hello(a test.TestArgs, r *test.TestResponse) error {
	//clustermanager.GetInstance()
	log.Info.Printf("Recibiendo saludo de %s.", a.NodeId)
	r.Result = true
	return nil
}
