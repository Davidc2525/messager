package main

import (
	"flag"
	"github.com/Davidc2525/messager/core/endpoint/websocketendpoint"
	"github.com/Davidc2525/messager/core/restapi"
	"github.com/Davidc2525/messager/services/new_user_conn"
	"github.com/Davidc2525/messager/services/rpc_connection_endpoint"
	"runtime"
	"strings"
	"time"

	"github.com/Davidc2525/messager/cluster/clustermanager"
	"github.com/Davidc2525/messager/cluster/server"
	"github.com/Davidc2525/messager/core/endpoint"
	"github.com/Davidc2525/messager/core/processor"
	"github.com/Davidc2525/messager/log"

	//"github.com/Davidc2525/messager/inst"
	"github.com/Davidc2525/messager/services/join"
	"github.com/Davidc2525/messager/services/ping"
	"github.com/Davidc2525/messager/services/test"
	"github.com/Davidc2525/messager/user"
	"github.com/Davidc2525/messager/userprovider"
	//"github.com/coreos/etcd"
)

var (
	log = mlog.New()
)

func main() {
	//fmt.Println(quote.Hello())
	log.Warning.Println("Num cpu", runtime.NumCPU(), runtime.GOMAXPROCS(2))
	var addrServer string
	var endPointAddr string
	var restapiaddr string
	var speers = ""
	var isSeed bool
	var seed string
	peers := []string{""}

	flag.StringVar(&addrServer, "addr", "localhost:2525", "Addr to server")
	flag.StringVar(&endPointAddr, "addrep", "localhost:2727", "Addr to endpint server (ws)")
	flag.StringVar(&speers, "peers", "", "Peers of cluster, string separate by comma (,): host:port,host2:port")
	flag.StringVar(&restapiaddr, "restapiaddr", "localhost:8080", "Addr for api rest server")
	flag.BoolVar(&isSeed, "isseed", false, "Spesified if this node is a server Seed")
	flag.StringVar(&seed, "seed", "", "Spesified the server seed")

	flag.Parse()

	//parse peers
	splitedpeers := strings.Split(speers, ",")
	for _, v := range splitedpeers {
		if v != "" {
			//host, port, err := net.SplitHostPort(strings.TrimSpace(v))
			peers = append(peers, v)

		}
	}
	log.Warning.Printf(addrServer, speers, peers)
	//os.Exit(1)

	wep := websocketendpoint.New(websocketendpoint.Conf{Addr: endPointAddr})
	eps := []endpoint.EndPointProvider{wep}

	apirestc := restapi.Conf{Addr: restapiaddr}
	pc := processor.Conf{
		EndPoints: eps,
		RestApi:   apirestc,
	}

	go processor.New(pc).Start()

	um := userprovider.GetInstance()

	um.AddUser(&user.User{Id: "55", Name: "nohe"})
	um.AddUser(&user.User{Id: "23034087", Name: "david"})
	um.AddUser(&user.User{Id: "123", Name: "luisa"})

	/*for _, v := range um.GetUsers() {
		j, _ := json.Marshal(v)
		var out bytes.Buffer
		json.Indent(&out, j, "", "\t")
		out.WriteTo(os.Stdout)
	}*/

	conf := &server.Conf{Addr: addrServer, Peers: peers, IsSeed: isSeed, Seed: seed}
	cm := clustermanager.GetInstance()
	rpc := new(rpc_connection_endpoint.RpcConectionEndPintService)
	//rpc_connection_endpoint.RemoteConnections =  make(map[string]map[string]processor.ConnectionEndPoint)
	go cm.StartAndJoin(conf,
		new(join.Join),
		new(ping.PingPong),
		new(test.Test),
		rpc,
		new(new_user_conn.NewConnUser))
	for {
		//fmt.Println(inst.Repository().Get("cm"))
		time.Sleep(time.Second * 2)

	}
	//server.New(conf)
}
