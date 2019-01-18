package new_user_conn

import (
	"github.com/Davidc2525/messager/log"
	t "github.com/Davidc2525/messager/services/new_user_conn/src"
	//"sync"
)

type NewConnUser struct {
	//mux sync.Mutex
}

var (
	log = mlog.New()
)

func (this *NewConnUser) NewConn(a t.Args, r *t.Response) error {
	//clustermanager.GetInstance()
	/*members :=clustermanager.GetInstance().GetMembersBy([]string{a.NodeId})
	members[0].GetRpc().Go("send",a,r,nil)
	processor.Pross.ConnIn <- processor.NewNewConn(NewRPCConnectionEndPoint())*/
	log.Info.Printf("Usuario conectado %s.", a.NodeId)
	r.Result = true
	return nil
}

func (this *NewConnUser) NewMessage(a t.Args, r *t.Response) error { //TODO
	return nil
}
