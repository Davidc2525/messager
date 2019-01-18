package rpc_connection_endpoint

import (
	"bytes"
	"encoding/gob"
	"github.com/Davidc2525/messager/core/endpoint"
	"github.com/Davidc2525/messager/core/endpoint/rpc_endpint"
	"github.com/Davidc2525/messager/core/packets"
	"github.com/Davidc2525/messager/core/processor"
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/services/rpc_connection_endpoint/src"
	"strings"
	"sync"
	"time"
)

//Schema: UID_CID = ConnectionEndPoint
type MapCOnn map[string]endpoint.ConnectionEndPoint

var (
	log    = mlog.New()
	lockEp sync.Mutex
	//RemoteConnections map[string]map[string]endpoint.ConnectionEndPoint

	Conns = make(MapCOnn)
)

func init() {
	go func() {
		for {
			time.Sleep(time.Second * 10)
			for k, v := range Conns {
				log.Warning.Printf("K: %v V: %#v", k, v)
			}
		}
	}()
}

func deleteAllPrefix(mapc MapCOnn, prefix string) {
	for k := range mapc {
		if strings.HasPrefix(k, prefix) {
			delete(mapc, k)
		}
	}
}

func newKey(host, uid, key string) string {
	if len(host) == 0 {
		return ""
	} else {
		if len(uid) > 0 {

			if len(key) > 0 {
				return host + "_" + uid + "_" + key
			}
			return host + "_" + "_" + uid
		} else {
			return host + "_"
		}
	}

	return ""
}

type RpcConectionEndPintService struct {
	//lockEp sync.Mutex
	//RemoteConnections map[string]map[string]endpoint.ConnectionEndPoint
}

func NewRpcConectionEndPintService() *RpcConectionEndPintService {

	RemoteConnections := make(map[string]map[string]endpoint.ConnectionEndPoint)
	rpc := new(RpcConectionEndPintService)
	rpc = &RpcConectionEndPintService{}
	RemoteConnections = RemoteConnections
	return rpc
}

func (this *RpcConectionEndPintService) NewRpcConnectionEp(args src.ArgsNewConnection, res *src.ResponseNewConnection) error {
	log.Warning.Printf("ADD %#v", args)
	lockEp.Lock()

	newConection := rpc_endpint.NewRPCConnectionEndPoint(args.Cid, args.Uid, args.Hostid)

	Conns[newKey(args.Hostid, args.Uid, args.Cid)] = newConection

	/*if conns,ok := RemoteConnections[args.Uid] ; ok {
		conns[args.Hostid] = newConection
	}else{
		nc :=  make(map[string]endpoint.ConnectionEndPoint)
		RemoteConnections[args.Uid] = nc
		RemoteConnections[args.Uid][args.Hostid] = newConection
	}*/
	lockEp.Unlock()
	processor.Pross.ConnEpIn <- newConection
	res.Connected = true

	return nil
}

func (this *RpcConectionEndPintService) RemoveRpcConnectionEp(args src.ArgsNewConnection, res *src.ResponseNewConnection) error {
	log.Warning.Printf("REMOVE %#v", args)
	lockEp.Lock()

	if conn, ok := Conns[newKey(args.Hostid, args.Uid, args.Cid)]; ok {
		delete(Conns, newKey(args.Hostid, args.Uid, args.Cid))
		conn.Close()
		processor.Pross.ConnEpOut <- conn
	}

	/*if conns,ok := RemoteConnections[args.Uid] ; ok {

		for _,c := range conns {
			delete(conns,c.GetId())

			processor.Pross.ConnEpOut <- c
		}
		RemoteConnections[args.Uid] = conns

	}*/
	lockEp.Unlock()
	res.Connected = true
	return nil
}

func (this *RpcConectionEndPintService) NewRpcConnectionPacket(args src.ArgsNewConnectionPacket, res *src.ResponseNewConnection) error {
	p := &packets.DContainer{}
	data := args.Packet
	gob.NewDecoder(bytes.NewBuffer(data)).Decode(p)
	//log.Warning.Printf("PACKET %#v %v", p,string(p.GetData()))
	//p.SetAttr("from","remote")

	//processor.Pross.MessageIn <- p
	processor.Pross.StreamIn <- p
	/*for _,c :=  range RemoteConnections[p.To]  {
		c.Send() <- data
	}*/
	res.Connected = true
	return nil
}
