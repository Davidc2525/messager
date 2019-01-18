package processor

import (
	"bytes"
	"encoding/json"
	"github.com/Davidc2525/messager/cluster/clustermanager"
	"github.com/Davidc2525/messager/core/endpoint"
	"github.com/Davidc2525/messager/core/packets"
	"github.com/Davidc2525/messager/core/precensemanager"
	_ "github.com/Davidc2525/messager/core/precensemanager/providers/etcd_provider_flat"
	"github.com/Davidc2525/messager/core/restapi"
	_ "github.com/Davidc2525/messager/core/restapi/api/apiuser"
	"github.com/Davidc2525/messager/core/sessionmanager"
	_ "github.com/Davidc2525/messager/core/sessionmanager/providers/etcd"
	_ "github.com/Davidc2525/messager/core/sessionmanager/providers/memory"
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/services/rpc_connection_endpoint/src"
	"github.com/Davidc2525/messager/user"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
)

//Schema: UID_CID = endpoint.ConnectionEndPoint
type MapCOnn map[string]endpoint.ConnectionEndPoint

var (
	log   = mlog.New()
	rapi  *restapi.RestApi
	Pross *Processor
	//packetTypeReg, _ = regexp.Compile(`(?ims)^(.+)!(.+)`)
)

func deleteAllPrefix(mapc MapCOnn, prefix string) {

	for k := range mapc {
		if strings.HasPrefix(k, prefix) {
			delete(mapc, k)
		}
	}
}

//filtrar conexiones por prefijo uid_cid
func getByPrefix(m MapCOnn, prefix string) (conns []endpoint.ConnectionEndPoint) {
	for k, v := range m {
		if strings.HasPrefix(k, prefix) {
			conns = append(conns, v)
		}
	}
	return
}

func NewKey2(uid, key string) string {
	if len(uid) > 0 {
		if len(key) > 0 {
			return uid + "_" + key
		}
		return uid + "_"
	} else {
		return ""
	}

	return ""
}

func newKey(host, uid, key string) string {
	if len(host) == 0 {
		return ""
	} else {
		if len(uid) > 0 {

			if len(key) > 0 {
				return host + "_" + uid + "_" + key
			}
			return host + "_" + uid
		} else {
			return host + "_"
		}
	}

	return ""
}

type UserConn struct {
	User *user.User
	//Id de la conexion
	Id            string
	Host          string
	Connected     bool
	Login         bool
	StartingChan  chan bool
	DisconectChan chan bool

	MessageIn  chan packets.MessagePacket
	PrecenseIn chan packets.PrecensePacket
	open       bool
	Conn       endpoint.ConnectionEndPoint
}

func NewUserConn(cep endpoint.ConnectionEndPoint) *UserConn {
	uc := &UserConn{Conn: cep}
	uc.DisconectChan = make(chan bool)
	uc.MessageIn = make(chan packets.MessagePacket, 10000)
	uc.PrecenseIn = make(chan packets.PrecensePacket)
	return uc
}

type NewConn struct {
	Cep endpoint.ConnectionEndPoint
}

func NewNewConn(cep endpoint.ConnectionEndPoint) *NewConn {
	return &NewConn{Cep: cep}
}

type Processor struct {
	InChan          chan string                                       //borrar
	UsersConn       map[string]map[string]*UserConn                   //borrar
	UsersConnEp     map[string]map[string]endpoint.ConnectionEndPoint //borrar
	UsersConnEpFlat MapCOnn
	UsersConnMux    sync.Mutex                           //borrar
	EndPoints       map[string]endpoint.EndPointProvider //borar

	StreamIn chan packets.Container

	MessageIn chan packets.Container
	EventIn   chan packets.Container

	PrecenseIn chan packets.PrecensePacket //borrar
	UserIn     chan packets.UserPacket     //borrar
	UserOut    chan packets.UserPacket     //borrar

	//conexiones entrantes
	ConnIn chan *NewConn //borrar

	ConnEpIn  chan endpoint.ConnectionEndPoint
	ConnEpOut chan endpoint.ConnectionEndPoint

	Conf        Conf
	EndPointMux sync.Mutex

	workers []*worker
}

type Conf struct {
	EndPoints []endpoint.EndPointProvider
	RestApi   restapi.Conf
}

func New(conf Conf) *Processor {

	log.Info.Println("Creating new Processor")
	var p = &Processor{}
	p.UsersConn = make(map[string]map[string]*UserConn)
	p.InChan = make(chan string)
	p.EndPoints = make(map[string]endpoint.EndPointProvider)
	p.UserIn = make(chan packets.UserPacket)
	p.UserOut = make(chan packets.UserPacket)
	p.ConnIn = make(chan *NewConn)
	p.MessageIn = make(chan packets.Container, 100000)
	p.EventIn = make(chan packets.Container, 100000)

	p.StreamIn = make(chan packets.Container, 100000)

	p.UsersConnEpFlat = make(MapCOnn)
	p.UsersConnEp = make(map[string]map[string]endpoint.ConnectionEndPoint)
	p.ConnEpIn = make(chan endpoint.ConnectionEndPoint)
	p.ConnEpOut = make(chan endpoint.ConnectionEndPoint)
	p.workers = []*worker{}
	p.Conf = conf
	rapi = restapi.New(conf.RestApi)

	Pross = p
	return p
}

/*func (this *Processor) handleConn(conn *NewConn) {
	defer func() {
		if v := recover(); v != nil {
			log.Warning.Println(v)
		}
	}()
	uc := NewUserConn(conn.Cep)

	//uc.Conn.Start()
	uc.open = false
	uc.Conn.SetHandleClose(func() {
		uc.open = false
		uc.DisconectChan <- true
	})
	uc.Conn.Start()
	if uc.Conn.GetWriter() == nil {

	} else {

	}

	if sess, err := sessionmanager.GlobalSessions.GetSession(uc.Conn.GetWriter(), uc.Conn.GetRequest(), false); err == nil {
		uid := sess.Get("uid").(string)
		if userStore, userErr := userprovider.GetInstance().GetUserById(uid); userErr == nil {
			var id = ksuid.New().String()
			hostId := clustermanager.GetInstance().Server.GetConf().Addr

			uc.User = userStore
			uc.Host = hostId
			uc.Id = id

			if perr := precensemanager.ProcenceManager.Provider.Add(uc.User.Id, hostId+"_"+id); perr == nil {

				uc.open = true

				this.UsersConnMux.Lock()
				if conns, ok := this.UsersConn[uc.User.Id]; ok {
					conns[uc.Id] = uc
				} else { // si no esta agregar al map
					this.UsersConn[uc.User.Id] = make(map[string]*UserConn)
					this.UsersConn[uc.User.Id][uc.Id] = uc
				}
				this.UsersConnMux.Unlock()

				go func() {
					for {
						select {
						case <-uc.DisconectChan:
							uc.open = false
							close(uc.DisconectChan)
							close(uc.MessageIn)
							close(uc.PrecenseIn)
							log.Warning.Println("Señal para desconectar", id)
							this.UsersConnMux.Lock()
							if conns, ok := this.UsersConn[uc.User.Id]; ok {
								if _, cok := conns[uc.Id]; cok {
									delete(conns, uc.Id)
								}
							}
							this.UsersConnMux.Unlock()

							if perr := precensemanager.ProcenceManager.Provider.Remove(uc.User.Id, hostId+"_"+id); perr != nil {
								log.Error.Printf("REMOVE %v", perr)
							}
							return

						case message := <-uc.MessageIn:
							data := bytes.Buffer{}
							json.NewEncoder(&data).Encode(message)
							uc.Conn.Send() <- data.Bytes()
						}
					}
				}()

				for uc.open { //read web socket
					mp := packets.NewDMessagePacket()
					select {
					case m, ok := <-uc.Conn.Receive():
						//si ok == false es por q la corrutina de websocket ya cerro el canal
						//per aun no ha sido llamada la funcion en HandleClose, por lo tanto this.open sigue siendo true
						//para cuando se comprueba q ok == false this.open se pasa a false y se sale del bucle
						//y HandleClose es llamada y cierra los demas canales
						if !ok {

							log.Error.Println("no ok", ok)
							//uc.DisconectChan <- true
							uc.open = false
							return
						}
						if len(m) > 0 {
							if m[1] == 33 { // 33 == !
								if m[0] == 109 { // 109 == m
									data := bytes.NewReader(m[2:])
									jd := json.NewDecoder(bufio.NewReader(data))
									jd.Decode(mp)

									this.MessageIn <- mp
									//uc.MessageIn <- mp
								}
							}
						}
					}

				}

			} else {
				log.Error.Printf("ADD %v", perr)
				uc.Conn.Close()
			}
		} else {
			uc.Conn.Close()
		}
	} else {
		uc.Conn.Close()
	}

}*/
func IsInstanceOf(objectPtr, typePtr interface{}) bool {
	return reflect.TypeOf(objectPtr) == reflect.TypeOf(typePtr)
}

type task struct {
	job func()
}

func (this *task) do(w *worker) {
	this.job()
}

type worker struct {
	id uint32
	//QueueTasks *list.List
	Tasks chan task
}

func (this *worker) start() {
	defer func() {
		if err := recover(); err != nil {
			log.Error.Println("error en loop: ", err)
		}
	}()

	for {
		select {
		case task, ok := <-this.Tasks:
			if !ok {
				return
			}
			log.Warning.Println("Recibida nueva tarea ", this.id, runtime.NumGoroutine())
			//
			//this.QueueTasks.PushBack(task)
			task.do(this)
		}
	}

}

func newWorker(id uint32) *worker {
	w := &worker{}
	w.id = id //ksuid.New().String()
	w.Tasks = make(chan task, 100000)
	//w.QueueTasks = list.New()
	go w.start()
	return w

}

var next int32 = 0

func (this *Processor) getWorker() *worker {

	w := this.workers[next]

	if next+1 >= int32(len(this.workers)) {
		next = 0
	} else {
		atomic.AddInt32(&next, 1)
	}
	return w
}

func (this *Processor) Start() {
	for k := range make([]bool, 100, 100) {
		w := newWorker(uint32(k))
		//log.Info.Println("append worker: ", w.id)
		this.workers = append(this.workers, w)
	}
	go func() {
		defer log.Error.Fatal("saliendo del bucle de canales")
		for {

			select {
			/*default:
			log.Info.Println(runtime.NumGoroutine())*/
			case conn := <-this.ConnIn:
				log.Warning.Println("		NEW CONNECTION")
				log.Info.Println(conn)
			//go this.handleConn(conn)
			case container, _ := <-this.StreamIn: //route

				stream := container.GetData()
				if len(stream) > 0 {
					if stream[1] == 33 { // 33 == !
						switch stream[0] {
						case 109: // 109 == message

							this.MessageIn <- container
						case 101: // 101 = event

							this.EventIn <- container
						}

					}
				}
			case container, _ := <-this.EventIn:
				t := func(container packets.Container) func() {
					return func() {
						this.UsersConnMux.Lock()
						defer this.UsersConnMux.Unlock()
						event := packets.NewDEventPacket()
						datap := container.GetData()
						ctype := datap[0:2]
						datain := datap[2:]
						datareader := bytes.NewReader(datain)

						json.NewDecoder(datareader).Decode(event)

						//log.Warning.Printf("	NEW EVENT %#v %v", event, ok)

						//this.UsersConnMux.Lock()
						//defer this.UsersConnMux.Unlock()
						//hostid := clustermanager.GetInstance().Server.GetConf().Addr

						connsBy := getByPrefix(this.UsersConnEpFlat, NewKey2(event.GetBy(), "")) //uid_
						connsTo := getByPrefix(this.UsersConnEpFlat, NewKey2(event.GetTo(), "")) //uid_
						if event.GetBy() != event.GetTo() {                                      //TODO add lock

							if container.GetClass() == packets.Remote { //remote

								if len(event.GetAttr("forward")) > 0 { //is fordward
									for _, c := range connsBy {
										if c.GetId() == container.GetCid() && c.Kind() == packets.Local { //send fordward to local
											data := bytes.Buffer{}
											event.SetAttr("forward", "true")
											json.NewEncoder(&data).Encode(event)
											c.Send() <- data.Bytes()
										}
									}
								} else { // no fordward
									for _, c := range connsTo {
										if c.Kind() == packets.Local && c.GetId() == container.GetCid() {
											data := bytes.Buffer{}
											event.SetAttr("forward", "true")
											json.NewEncoder(&data).Encode(event)
											c.Send() <- data.Bytes()
										}
									}
								}

							} else { //local
								for _, c := range connsTo {
									data := bytes.Buffer{}

									json.NewEncoder(&data).Encode(event)
									if c.Kind() == packets.Remote {
										dataSend := []byte(string(ctype) + string(data.Bytes()))
										//log.Warning.Printf("	NEW EVENT SENDING TO REMOTE %v", string(dataSend))
										c.Send() <- dataSend
									} else {
										c.Send() <- data.Bytes()
									}
								}

								for _, c := range connsBy { //fordwarding
									if c.Kind() == packets.Remote {
										data := bytes.Buffer{}
										//data.Write([]byte("e!"))
										event.SetAttr("forward", "true")
										json.NewEncoder(&data).Encode(event)
										dataSend := []byte(string(ctype) + string(data.Bytes()))
										//log.Warning.Printf("	NEW EVENT SENDING TO REMOTE %v", string(dataSend))
										c.Send() <- dataSend
									}
								}

							}

						}
					}
				}(container)

				this.getWorker().Tasks <- task{job: t}

			case container, ok := <-this.MessageIn:
				//TODO hande new messaje
				if !ok {
					return
				}
				t := func(container packets.Container) func() {
					return func() {
						message := packets.NewDMessagePacket()
						datap := container.GetData()
						ctype := datap[0:2]
						datain := datap[2:]
						datareader := bytes.NewReader(datain)

						json.NewDecoder(datareader).Decode(message)
						//log.Warning.Printf("		NEW MESSAGE %v %#v", runtime.NumGoroutine(), message)
						this.UsersConnMux.Lock()
						defer this.UsersConnMux.Unlock()
						//hostid := clustermanager.GetInstance().Server.GetConf().Addr

						connsBy := getByPrefix(this.UsersConnEpFlat, NewKey2(message.GetBy(), "")) //uid_
						connsTo := getByPrefix(this.UsersConnEpFlat, NewKey2(message.GetTo(), "")) //uid_

						if message.GetBy() != message.GetTo() { //TODO add lock

							if container.GetClass() == packets.Remote { //remote

								if len(message.GetAttr("forward")) > 0 { //is fordward
									for _, c := range connsBy {
										if c.GetId() == container.GetCid() && c.Kind() == packets.Local { //send fordward to local
											data := bytes.Buffer{}
											message.SetAttr("forward", "true")
											json.NewEncoder(&data).Encode(message)
											c.Send() <- data.Bytes()
										}
									}
								} else { // no fordward
									for _, c := range connsTo {
										if c.Kind() == packets.Local && c.GetId() == container.GetCid() {
											data := bytes.Buffer{}
											message.SetAttr("forward", "true")
											json.NewEncoder(&data).Encode(message)
											c.Send() <- data.Bytes()
										}
									}
								}

							} else { //local
								for _, c := range connsTo {
									data := bytes.Buffer{}
									json.NewEncoder(&data).Encode(message)
									if c.Kind() == packets.Remote {
										//ctype + payload
										// m! + {...}
										// m!{...}
										c.Send() <- []byte(string(ctype) + string(data.Bytes()))
									} else {
										c.Send() <- data.Bytes()
									}

								}

								for _, c := range connsBy { //fordwarding
									data := bytes.Buffer{}
									message.SetAttr("forward", "true")
									json.NewEncoder(&data).Encode(message)
									if c.Kind() == packets.Remote {
										c.Send() <- []byte(string(ctype) + string(data.Bytes()))
									} else {
										c.Send() <- data.Bytes()
									}
								}

							}

						}

						/*if message.GetBy() != message.GetTo() { //TODO add lock
							for _, c := range getByPrefix(this.UsersConnEpFlat, NewKey2(message.GetTo(), "")) {
								if message.GetAttr("from") == "local" {
									data := bytes.Buffer{}
									json.NewEncoder(&data).Encode(message)
									if c.IsOpen() {
										c.Send() <- data.Bytes()
									}
								}

								if message.GetAttr("from") == "remote" {
									if c.GetKind() == "local" {
										if c.GetId() == message.GetAttr("cid") {
											data := bytes.Buffer{}
											json.NewEncoder(&data).Encode(message)
											if c.IsOpen() {
												c.Send() <- data.Bytes()
											}
										}
									}
								}
							}

						}

						//FordWarding
						for _, c := range getByPrefix(this.UsersConnEpFlat, NewKey2(message.GetBy(), "")) {
							if c.IsOpen() {
								if c.GetKind() == "local" {
									data := bytes.Buffer{}
									message.SetAttr("forward", "true")
									json.NewEncoder(&data).Encode(message)
									c.Send() <- data.Bytes()
								}

							}
							//log.Warning.Printf("		FORWARDED TO %v %#v", message.GetBy(), message)
						}*/
					}
				}(container)

				this.getWorker().Tasks <- task{job: t}

			case cep := <-this.ConnEpIn:
				t := func(cep endpoint.ConnectionEndPoint) func() {
					return func() {
						this.UsersConnMux.Lock()
						uid := cep.GetUid()
						cid := cep.GetId()
						//hostid := clustermanager.GetInstance().Server.GetConf().Addr

						this.UsersConnEpFlat[NewKey2(uid, cid)] = cep

						log.Warning.Printf("		ADD NEW CONNECTION %v %v", uid, cep.GetId())
						if false {
							if conns, ok := this.UsersConnEp[uid]; ok {
								conns[cep.GetId()] = cep
							} else { // si no esta agregar al map
								this.UsersConnEp[uid] = make(map[string]endpoint.ConnectionEndPoint)
								this.UsersConnEp[uid][cep.GetId()] = cep
							}
						}

						if cep.Kind() != packets.Remote {
							//log.Warning.Printf("		SENDING RPC %v %v", uid, cep.GetId())
							//hid := clustermanager.GetInstance().Server.Id
							for _, m := range clustermanager.GetInstance().GetMembersBy(clustermanager.GetInstance().GetMembersIds()) {
								args := src.ArgsNewConnection{Uid: cep.GetUid(), Cid: cep.GetId(), Hostid: cep.Host()}
								respo := src.ResponseNewConnection{}
								m.GetRpc().Go("RpcConectionEndPintService.NewRpcConnectionEp", args, &respo, nil)
							}
						}
						this.UsersConnMux.Unlock()
						log.Warning.Printf("		ADDED NEW CONNECTION %v %v", uid, cep.GetId())
					}
				}(cep)
				this.getWorker().Tasks <- task{job: t}

			case cep := <-this.ConnEpOut:
				t := func(cep endpoint.ConnectionEndPoint) func() {
					return func() {

						this.UsersConnMux.Lock()
						uid := cep.GetUid()
						cid := cep.GetId()
						//hostid := clustermanager.GetInstance().Server.Id

						log.Warning.Printf("		DEL NEW CONNECTION %v %v", uid, cep.GetId())

						if _, ok := this.UsersConnEpFlat[NewKey2(uid, cid)]; ok {
							delete(this.UsersConnEpFlat, NewKey2(uid, cid))
							//Pross.ConnEpOut <- conn
						}

						if false {
							if conns, ok := this.UsersConnEp[uid]; ok {
								if _, cok := conns[cep.GetId()]; cok {
									delete(conns, cep.GetId())
								}
							}
						}

						if cep.Kind() != packets.Remote {
							log.Warning.Printf("		SENDING RPC %v %v", uid, cep.GetId())

							for _, m := range clustermanager.GetInstance().GetMembersBy(clustermanager.GetInstance().GetMembersIds()) {
								args := src.ArgsNewConnection{Uid: cep.GetUid(), Cid: cep.GetId(), Hostid: cep.Host()}
								respo := src.ResponseNewConnection{}
								m.GetRpc().Go("RpcConectionEndPintService.RemoveRpcConnectionEp", args, &respo, nil)
							}
						}
						this.UsersConnMux.Unlock()
						log.Warning.Printf("		DELETED NEW CONNECTION %v %v", uid, cep.GetId())

						/*case nuser := <-this.UserIn:
							log.Warning.Println("		USER CONECT")
							log.Info.Println(nuser.(packets.UserPacket))
						case outuser := <-this.UserOut:
							log.Warning.Println("		USER DISCONECT")
							log.Info.Println(outuser.(packets.UserPacket))*/
					}
				}(cep)
				this.getWorker().Tasks <- task{job: t}
			}
		}
	}()
	//log.Printf("%#v",this.Conf.EndPoints)
	//log.Println("add end point",k,ep.GetName())
	//memory.GetProvider()
	precensemanager.New("etcd_flat")
	sessionmanager.GetInstance("etcd", "gosessionid", 3600*24*365)
	rapi.Start()

	for _, ep := range this.Conf.EndPoints {
		this.AddEndPoint(ep)
		if err := ep.Start(); err != nil {
			log.Error.Println(err)
		}
	}
	go func() {
		for message := range this.InChan {

			log.Info.Println(message)
		}
	}()

}

func (this *Processor) AddEndPoint(ep endpoint.EndPointProvider) {
	this.EndPointMux.Lock()
	var name = ep.GetName()
	log.Info.Println("Add End Point", name)

	if _, p := this.EndPoints[name]; !p {
		this.EndPoints[name] = ep
	}
	this.EndPointMux.Unlock()
}

func Ints(input []endpoint.ConnectionEndPoint) []endpoint.ConnectionEndPoint {
	u := make([]endpoint.ConnectionEndPoint, 0, len(input))
	m := make(map[endpoint.ConnectionEndPoint]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}

	return u
}
