package join

import (
	"github.com/Davidc2525/messager/cluster/clustermanager"
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/services/join/src"

	"net/rpc"
	//"time"
)

type Join struct {
	S *rpc.Server
	//mux sync.Mutex
}

var (
	log = mlog.New()
)

func (this *Join) JoinCluster(a join.JoinArgs, r *join.JoinResponse) error {
	cm := clustermanager.GetInstance()
	ac := false //colocar en false si solo se aceptaran los miembros q son pasados en peers
	log.Info.Println("joining client", a, *r)
	if cm.Mode == clustermanager.CLUSTER {
		for _, p := range cm.Conf.Peers {
			if p == a.NodeId {

				log.Info.Println("Accept to join node: ", a.NodeId)
				ac = true
			}
		}
	}
	r.Result = ac
	r.RemoteHostId = cm.Server.Id

	/*go func() {
		if !cm.HasMember(a.NodeId) {
			time.Sleep(time.Second * 5)
			log.Info.Println("not a member in local, addiding")
			go cm.ConectWithMember(a.NodeId)
		}
	}()*/

	return nil
}
