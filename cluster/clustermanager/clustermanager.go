package clustermanager

import (
	"github.com/Davidc2525/messager/log"
	"github.com/hashicorp/memberlist"
	"net"
	"strings"

	//"context"
	//ctx "github.com/Davidc2525/messager/cluster/context"
	"fmt"
	"github.com/Davidc2525/messager/cluster/client"

	//"github.com/Davidc2525/messager/cluster/member"
	"github.com/Davidc2525/messager/cluster/server"
	"github.com/Davidc2525/messager/inst"

	//"github.com/Davidc2525/messager/services/join"

	"sync"
	"time"

	"github.com/Davidc2525/messager/services/test/src"
)

const (
	STAND = iota + 1
	CLUSTER
)

var (
	log = mlog.New()
)
var instance *ClusterManager

type ClusterManager struct {
	Server    *server.Server
	Members   map[string]*client.Client
	memberIn  chan client.Client
	memberOut chan client.Client

	Mode int

	Conf *Conf

	discoveryIn  chan string
	discoveryOut chan string

	mux sync.Mutex
}

type Event struct {
}

func (this *ClusterManager) NotifyJoin(node *memberlist.Node) {
	fmt.Println("nuevo node", node, node.Addr)
	this.discoveryIn <- node.Addr.To4().String()

	//this.EventDelegate.NotifyJoin(node)

}
func (this *ClusterManager) NotifyLeave(node *memberlist.Node) {
	fmt.Println("bay node", node, node.Addr)
	//this.EventDelegate.NotifyJoin(node)
	this.discoveryOut <- node.Addr.To4().String()
}
func (this *ClusterManager) NotifyUpdate(node *memberlist.Node) {
	fmt.Println("update node", node)
	//this.EventDelegate.NotifyJoin(node)

}

func GetInstance() *ClusterManager {
	/*go func() {
		list, err := memberlist.Create(memberlist.DefaultLocalConfig())
		if err != nil {
			panic("Failed to create memberlist: " + err.Error())
		}

		// Join an existing cluster by specifying at least one known member.
		n, err := list.Join([]string{"1.2.3.4"})
		if err != nil {
			panic("Failed to join cluster: " + err.Error())
		}

		// Ask for members of the cluster
		for _, member := range list.Members() {
			fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
		}
	}()*/
	if instance == nil {
		instance = &ClusterManager{discoveryIn: make(chan string), discoveryOut: make(chan string), memberOut: make(chan client.Client), memberIn: make(chan client.Client), Members: make(map[string]*client.Client)}
	}
	inst.Repository().Set("cm", instance)
	return instance
}

type Conf struct {
	Peers      []string
	ServerConf *server.Conf
	ClientConf *client.Conf
}

//Unirse al cluster, crear un server y se conecta con los demas
func (this *ClusterManager) StartAndJoin(conf *Conf, services ...interface{}) {

	this.Server = server.New(conf.ServerConf)
	this.Conf = conf
	if len(conf.Peers) > 0{
		this.Mode = CLUSTER
	}else{
		this.Mode = STAND
	}
	go func() {
		for {
			time.Sleep(time.Second * 10)
			for _, v := range this.Members {
				log.Info.Println("member ", v)
			}
		}
	}()

	go func() {
		return
		list, err := memberlist.Create(memberlist.DefaultLocalConfig())
		if err != nil {
			panic("Failed to create memberlist: " + err.Error())
		}

		if !this.Server.GetConf().IsSeed {

			// Join an existing cluster by specifying at least one known member.
			seed := ""
			h, _, _ := net.SplitHostPort(this.Server.GetConf().Seed)
			seed = h
			n, err := list.Join([]string{seed})
			if err != nil {
				panic("Failed to join cluster: " + err.Error())
			}
			if n > 0 {
			}
		}

		// Ask for members of the cluster
		for _, member := range list.Members() {
			fmt.Printf("Member: %s %s\n", member.Name, member.Addr)
		}

	}()

	go func() {
		if len(this.Conf.Peers) == 0 {
			return
		}

		log.Info.Println("waiting for members")
		for {

			select {
			case nm := <-this.memberIn:
				fmt.Println("new member", nm.Conf)
				this.Members[nm.Conf.Addr] = &nm
			case nm := <-this.memberOut:
				fmt.Println("member bay", nm.Conf)
				delete(this.Members, nm.Conf.Addr)

				//TODO tratar de conectar con ese cliente de nuevo

				go this.ConectWithMember(nm.Conf.Addr)
			}
		}

	}()

	for k, v := range services {
		log.Info.Println("add Service", k, v)
		this.Server.AddService(v)
	}
	//server.AddService(this)
	go this.Server.Start()
	//go this.helloMembers()

	go this.waitForMemberDiscovery()
	go this.conectWithMembers()

	this.Server.Join()
}

func (this *ClusterManager) waitForMemberDiscovery() { //escuchar peers para conectar

	for {
		select {
		case discovered := <-this.discoveryOut:
			delete(this.Members, discovered+":2525")

		case discovered := <-this.discoveryIn:
			var addr, host, port string
			if strings.Index(discovered, ":") == -1 {
				discovered = discovered + ":"
			}

			discovered = strings.ToLower(discovered)

			log.Warning.Printf("		IN (%v)", discovered)

			dhost, dport, e := net.SplitHostPort(discovered)

			log.Warning.Printf("		DHOST (%v) DPORT (%v)", dhost, dport)

			if e == nil {
				if dhost != "" {
					host = dhost
					if dport != "" {
						port = dport
					} else {
						port = "2525"
					}
				} else {
					host = "localhost"
				}
				addr = net.JoinHostPort(host, port)
				log.Warning.Println("Input discovered ", addr)
				this.ConectWithMember(addr)
			}
		}

	}
}

func (this *ClusterManager) conectWithMembers() { //conectar directamente pasados por -peers
	log.Trace.Println("peers", this.Conf.Peers)
	if len(this.Conf.Peers) >= 1 {
		for _, v := range this.Conf.Peers {
			if v != this.Server.GetConf().Addr && v != "" {
				//this.ConectWithMember(v)
				this.discoveryIn <- v
			}
		}
	}
}
func (this *ClusterManager) ConectWithMember(addr string) {
	cConf := client.Conf{}

	cConf.ServerAddr = this.Server.GetConf().Addr
	cConf.ServerId = this.Server.Id
	cConf.Addr = addr
	cConf.Retry = this.Conf.ClientConf.Retry
	cConf.RetryCount = this.Conf.ClientConf.RetryCount
	cConf.RetryDelay = this.Conf.ClientConf.RetryDelay

	go client.New(cConf, this.memberIn, this.memberOut)
}

func (this *ClusterManager) AddMember() {

}

func (this *ClusterManager) RemoveMember() {

}

func (this *ClusterManager) HasMember(id string) (p bool) {
	_, p = this.Members[id]
	return p
}

//obtener los ids de los miembros
func (this *ClusterManager) GetMembersIds() (peers []string) {
	for k := range this.Members {
		peers = append(peers, k)
	}

	return
}

//obtener los ids de los miembros
func (this *ClusterManager) GetMembers() (mebs map[string]*client.Client) {
	return this.Members
}

func (this *ClusterManager) GetMembersBy(members []string) (mebs []*client.Client) {
	for _, member := range this.Members {
		if indexOf(members, member.Conf.Addr) != -1 {
			mebs = append(mebs, member)
		}
	}
	return

}

func indexOf(slice []string, key string) int {
	var p = -1
	for k, v := range slice {
		if v == key {
			p = k
			return p
		}
	}
	return p
}

func (this *ClusterManager) helloMembers() {
	for {
		time.Sleep(time.Millisecond * 10000)
		ms := this.GetMembers()
		if len(ms) > 0 {
			fmt.Println("Saludando miembros")

			for _, c := range ms {
				args := test.TestArgs{this.Server.GetConf().Addr}

				resp := test.TestResponse{}
				c.GetRpc().Call("Test.Hello", args, &resp)
			}
		}
	}
}

/*func (this *ClusterManager) JoinCluster(a join.JoinArgs, r *join.JoinResponse) error {
	//clustermanager.GetInstance()
	ac := false
	fmt.Println("joining client", a, *r)
	for _, p := range this.Server.GetConf().Peers {
		if p == a.NodeId {

			fmt.Println("Accept to join node: ", a.NodeId)
			ac = true
		}
	}
	r.Result = ac

	return nil
}*/
