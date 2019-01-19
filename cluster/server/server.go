package server

import (
	"fmt"
	"github.com/Davidc2525/messager/log"
	"github.com/segmentio/ksuid"

	//"github.com/Davidc2525/messager/cluster/client"

	"net"
	"net/rpc"
	"sync"
	"time"
)

var (
	log = mlog.New()
)

type Conf struct {
	Addr  string
	//Peers []string

	Seed   string
	IsSeed bool
}

//Serer
type Server struct {
	Id        string
	listener  net.Listener
	rpcServer *rpc.Server
	clients   map[string]*ClientConn
	members   map[string]*rpc.Client
	conf      *Conf
	done      chan bool
	mux       sync.Mutex
}

type ClientConn struct {
	Conn *net.Conn
}

func (this *Server) GetConf() *Conf {
	return this.conf
}

func (this *Server) Join() {
	for {
		time.Sleep(time.Second * 60)
	}
}

func New(conf *Conf) *Server {

	var server *Server = &Server{clients: make(map[string]*ClientConn)}
	server.Id = ksuid.New().String()
	var Defaultaddr string = "localhost:2525"

	if len(conf.Addr) == 0 {
		conf.Addr = Defaultaddr
	}
	server.conf = conf

	//server.rpcServer = rpc.NewServer()

	/*err := server.start()
	fmt.Println("after start")
	if err == nil {
		for _, v := range server.conf.Peers {
			cConf := client.Conf{Addr: v, Retry: true, RetryCount: 1000, RetryDelay: 1}
			go client.New(&cConf, members)
		}

		defer func() {
			server.done <- true
		}()

	}*/

	return server
}

func (this *Server) AddService(service interface{}) {
	log.Info.Println("AddService", service)
	rpc.Register(service)
}

func (this *Server) Start() error {
	/*members := make(chan client.Client)

	go func() {
		log.Println("waiting for members")
		for {
			select {
			case nm := <-members:
				fmt.Println("new member", *nm.Conf)
			}
		}
	}()*/

	listener, _ := net.Listen("tcp", this.conf.Addr)
	rpc.Accept(listener)

	return nil
}

func (this *Server) accept() {
	for {
		/*
			 select {
			case <-this.done:
				fmt.Println("server detenido")
				return
			}*/

		newConn, err := this.listener.Accept()
		rpc.NewClient(newConn)
		if err != nil {
			fmt.Println("Error al aceptar nueva conexion")
		} else {
			go this.handleConn(ClientConn{&newConn})
		}
	}
}

func (this *Server) handleConn(client ClientConn) {
	//this.addClient(client)
	this.rpcServer.ServeConn(*client.Conn)
	//this.removeClient(client)
}

func (this *Server) addClient(client ClientConn) {
	this.mux.Lock()
	var id string = (*client.Conn).RemoteAddr().String()
	fmt.Println("New Client", id, client)
	this.clients[id] = &client
	this.mux.Unlock()
}
func (this *Server) removeClient(client ClientConn) {
	this.mux.Lock()
	var id string = (*client.Conn).RemoteAddr().String()
	fmt.Println("Client disconnected", id, client)
	delete(this.clients, id)
	this.mux.Unlock()
}

func (this *Server) Stot() {
	this.done <- true
}

/*------------*/
type Join struct {
	S *rpc.Server
	//mux sync.Mutex
}

type JoinArgs struct {
	NodeId string
}

type JoinResponse struct {
	Result bool
}

func (this *Join) JoinCluster(a *JoinArgs, r *JoinResponse) error {
	//clustermanager.GetInstance()
	r.Result = true
	return nil
}
