package rpc_endpint

import (
	"bytes"
	"encoding/gob"
	"github.com/Davidc2525/messager/cluster/clustermanager"
	"github.com/Davidc2525/messager/core/endpoint"
	"github.com/Davidc2525/messager/core/packets"
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/services/rpc_connection_endpoint/src"
	"net/http"
)

import (
	"github.com/Davidc2525/messager/cluster/client"
)

var (
	log = mlog.New()
)

type RPCConnectionEndPoint struct {
	id     string
	uid    string
	hostId string

	con *client.Client

	open bool

	MessageIn chan []byte
	//MessageOut chan []byte

	onOpen        chan bool
	onClose       chan bool
	handleOnCLose func()
}

func (this *RPCConnectionEndPoint) Host() string {
	return this.hostId
}

func (this *RPCConnectionEndPoint) Kind() packets.ClassConnection {
	return packets.Remote
}

func (this *RPCConnectionEndPoint) IsOpen() bool {
	//panic("implement me")

	return this.open

}

func NewRPCConnectionEndPoint(cid, uid, hostid string) endpoint.ConnectionEndPoint {
	rpcep := &RPCConnectionEndPoint{id: cid, uid: uid, hostId: hostid}
	rpcep.MessageIn = make(chan []byte, 10000)
	//rpcep.MessageOut = make(chan []byte)
	rpcep.onOpen = make(chan bool)
	rpcep.onClose = make(chan bool)
	rpcep.open = true
	rpcep.handleOnCLose = func() {

		log.Error.Println("Cerando rpc endpoint", cid, uid, hostid)
		rpcep.open = false
		close(rpcep.MessageIn)
		close(rpcep.onClose)
		close(rpcep.onOpen)
	}
	go rpcep.Start()
	return rpcep
}

func (this *RPCConnectionEndPoint) GetUid() string {
	//panic("implement me")
	return this.uid
}

func (this *RPCConnectionEndPoint) GetId() string {
	//panic("implement me")
	return this.id
}

func (this *RPCConnectionEndPoint) Send() chan<- []byte {
	//panic("implement me")
	return this.MessageIn
}

func (this *RPCConnectionEndPoint) OnClose() <-chan bool {
	//panic("implement me")
	return nil
}

func (this *RPCConnectionEndPoint) OnOpen() <-chan bool {
	//panic("implement me")
	return nil
}

func (this *RPCConnectionEndPoint) Close() {
	this.handleOnCLose()
}

func (this *RPCConnectionEndPoint) Start() {
	//panic("implement me")
	go func() {

		for this.open {
			select {
			case v, ok := <-this.MessageIn:
				if !ok {
					this.open = false
					return
				}
				//og.Warning.Printf("\nRPC ENDPOINT  data %v",string(v))
				container := packets.NewDContainer(packets.Remote, this.hostId, false, v)
				container.Cid = this.id
				//container.Class = packets.Remote
				//container.Data = v

				dataSend := bytes.NewBuffer([]byte{})
				gob.NewEncoder(dataSend).Encode(container)

				//log.Warning.Printf("\n\nRPC ENDPOINT  %#v %v",container,string(v))

				for _, con := range clustermanager.GetInstance().GetMembersBy([]string{this.hostId}) {
					args := src.ArgsNewConnectionPacket{Packet: dataSend.Bytes()}
					respo := src.ResponseNewConnection{}
					con.GetRpc().Go("RpcConectionEndPintService.NewRpcConnectionPacket", args, &respo, nil)
				}
			}

		}

	}()
}

func (this *RPCConnectionEndPoint) GetWriter() http.ResponseWriter {
	return nil
}

func (this *RPCConnectionEndPoint) GetRequest() *http.Request {
	return nil
}

func (this *RPCConnectionEndPoint) SetHandleClose(func()) {
	//panic("implement me")
}
