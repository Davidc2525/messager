package client

import (
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/services/join/src"
	pingpong "github.com/Davidc2525/messager/services/ping/src"
	//"net"
	"net/rpc"
	"time"
)

var (
	log = mlog.New()
)

// miembro del cluster
type Client struct {
	Id         string // toma el id del server remoto
	rpcclient  *rpc.Client
	Conected   bool
	Joined     bool
	chconected chan bool
	chanJoin   chan bool
	Conf       Conf
}

type Conf struct {
	Addr       string
	ServerAddr string
	Retry      bool
	RetryCount int
	RetryDelay int
	ServerId   string
}

func New(conf Conf, chanClients chan Client, chanDisconect chan Client) {
	var client = new(Client)
	//client.Id = ksuid.New().String()
	client.chconected = make(chan bool)
	client.chanJoin = make(chan bool)
	log.Info.Println(conf)
	var defaultAddr = "localhost:2525"
	//var defaultRetry bool = false
	//var defaultRetryCount int = 10
	var defaultRetryDelay = 5

	if conf.Addr == "" {
		conf.Addr = defaultAddr
	}
	if conf.Retry {
		if conf.RetryDelay == 0 {
			conf.RetryDelay = defaultRetryDelay
		}
	}

	//escuchar canales para avisar a ClusterManager estado de cocexion
	go func() {
		log.Info.Println("waiting for conected")
		for {
			select {
			case cd := <-client.chanJoin:
				if cd {
					client.Joined = true
					chanClients <- *client

				}
				if !cd {
					client.Joined = false
					chanDisconect <- *client
				}
			}
		}
	}()

	client.Conf = conf
	client.tryConnect()

	args := join.JoinArgs{conf.ServerAddr}

	resp := join.JoinResponse{}

	log.Info.Println("sedin rpc join", conf.Addr)
	client.rpcclient.Call("Join.JoinCluster", args, &resp)
	log.Info.Println("response joining rpc", resp)
	if !resp.Result {
		log.Warning.Printf("Can´t join to cluster")
		client.rpcclient.Close()
		client.chanJoin <- false
	} else {
		client.Id = resp.RemoteHostId
		log.Info.Println("Joined", resp)
		go client.makePingPong() //TODO QUITAR
		client.chanJoin <- true

	}

}

func (this *Client) makePingPong() {
	pongFailds := 0
	time.Sleep(time.Second * 2)
	for {
		if !this.Joined {
			time.Sleep(time.Second)
			return
		}
		log.Info.Println("Haciando Ping")

		args := pingpong.PingArgs{NodeId: this.Conf.ServerAddr}

		resp := pingpong.PingResponse{}
		err := this.rpcclient.Call("PingPong.Ping", args, &resp)

		if err != nil {
			pongFailds++
			log.Info.Println("No se pudo recibir el pong")
			if pongFailds > 0 {
				log.Info.Println("El cliente se elminara")
				this.chanJoin <- false
				break
			}
		} else {
			if !resp.Result {
				log.Info.Println("No se acepto la union")
				this.chanJoin <- false
				break

			} else {

				log.Info.Println("Pong recibido")
			}
		}

		time.Sleep(time.Second * 5)

	}
}

func (this *Client) tryConnect() {
	var retrys = 0
	for {

		if retrys >= this.Conf.RetryCount {
			log.Info.Println("Can´t conect to server, no more retrys", this.Conf)
			return
		}
		if c := this.Conected; !c /*si no esta conectado*/ {
			log.Info.Println("tryConnect", this.Conf)
			//con, err := net.Dial("tcp", this.Conf.Addr)

			client, err := rpc.Dial("tcp", this.Conf.Addr)
			if err != nil {
				log.Warning.Println("dialing:", err)
				retrys++
			} else {
				//this.chconected <- true
				this.Conected = true
				this.rpcclient = client
				log.Info.Println("Connected", this.Conf)
				//log.Println("con", con.LocalAddr(), con.RemoteAddr())

			}

		} else {

			return
		}

		time.Sleep(time.Second * time.Duration(this.Conf.RetryDelay))
		if !this.Conf.Retry {
			return
		}
	}
}

func (this *Client) GetRpc() *rpc.Client {
	return this.rpcclient
}
