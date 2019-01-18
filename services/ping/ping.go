package ping

import (
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/services/ping/src"
	//"sync"
)

type PingPong struct {

	//mux sync.Mutex
}

var (
	log = mlog.New()
)

func (this *PingPong) Ping(a *ping.PingArgs, r *ping.PingResponse) error {
	//clustermanager.GetInstance()
	log.Info.Printf("Reciviendo Ping de %s, enviado Pong", a.NodeId)
	r.Result = true
	return nil
}

/*func (this *PingPong) Pong(a ping.PongArgs, r *ping.PongResponse) error {
	//clustermanager.GetInstance()
	r.Result = true
	return nil
}
**/
