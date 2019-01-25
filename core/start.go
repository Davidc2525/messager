package core

import (
	"flag"
	"github.com/Davidc2525/messager/cluster/client"
	"github.com/Davidc2525/messager/cluster/clustermanager"
	"github.com/Davidc2525/messager/cluster/server"
	"github.com/Davidc2525/messager/core/endpoint"
	"github.com/Davidc2525/messager/core/endpoint/websocketendpoint"
	"github.com/Davidc2525/messager/core/messagemanager"
	"github.com/Davidc2525/messager/core/messagemanager/provider/default_provider"
	"github.com/Davidc2525/messager/core/processor"
	"github.com/Davidc2525/messager/core/restapi"
	"github.com/Davidc2525/messager/core/user"
	"github.com/Davidc2525/messager/core/userprovider"
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/services/join"
	"github.com/Davidc2525/messager/services/new_user_conn"
	"github.com/Davidc2525/messager/services/ping"
	"github.com/Davidc2525/messager/services/rpc_connection_endpoint"
	"github.com/Davidc2525/messager/services/test"
	"runtime"
	"strings"
	"time"
)

var (
	log = mlog.New()
)

func Start() {

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
	clientConf := &client.Conf{Retry: true, RetryCount: 1000, RetryDelay: 1}
	conf := &server.Conf{Addr: addrServer, IsSeed: isSeed, Seed: seed}
	clusterConf := &clustermanager.Conf{Peers: peers, ServerConf: conf, ClientConf: clientConf}

	messagemanager.Register("memory", default_provider.NewDefaultProvider())
	messagemanager.InitManager("memory")

	cm := clustermanager.GetInstance()
	rpc := new(rpc_connection_endpoint.RpcConectionEndPintService)
	//rpc_connection_endpoint.RemoteConnections =  make(map[string]map[string]processor.ConnectionEndPoint)
	go cm.StartAndJoin(clusterConf,
		new(join.Join),
		new(ping.PingPong),
		new(test.Test),
		rpc,
		new(new_user_conn.NewConnUser))



	messagemanager.Manager.Pder.GetStore().CreateInbox(&user.User{Id:"123"},[]*user.User{})
	messagemanager.Manager.Pder.GetStore().CreateInbox(&user.User{Id:"55"},[]*user.User{})

	for {
		//fmt.Println(inst.Repository().Get("cm"))
		time.Sleep(time.Second * 2)

	}
}
