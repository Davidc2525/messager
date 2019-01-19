package messagemanager

import (
	"github.com/Davidc2525/messager/core/packets"
	"github.com/Davidc2525/messager/log"
	"github.com/Davidc2525/messager/user"
	"sync"
)

var (
	Manager  *MessageManager
	provides = make(map[string]Provider)
	log      = mlog.New()
	once     sync.Once
)

const (
	GETUSERS = iota
	UPDATEANDGET
	DELETE
)

type Provider interface {
	MakeOp() chan Op
}

type InfoPacket struct {
	Paced bool
	To    []*user.User
}

//paquete de proceso q contiene mensaje o evento producido,
//q tiene un canal para enviar informacion de vuelta
type ProcessPacket struct { //actor
	Receive chan *InfoPacket
	Packet  *packets.Packet
}

type MessageManager struct {
	Pder Provider

	//Recibe paquetes de proceso
	In chan *ProcessPacket
}

func InitManager(provider string) {
	if p, ok := provides[provider]; ok {
		once.Do(func() {
			Manager = &MessageManager{Pder: p, In: make(chan *ProcessPacket)}
		})
		go Manager.start()
	} else {

		log.Error.Println("no existe proveedor", provider)
		once.Do(func() {
			Manager = &MessageManager{In: make(chan *ProcessPacket)}
		})
		go Manager.start()
	}

}
func init() {
	//InitManager("memory")
}
func (this *MessageManager) start() {
	go func() {

		for {

			var p *ProcessPacket
			var ok bool

			select {
			case p, ok = <-this.In:
				if ok {
					switch v := (*p.Packet).(type) {
					case packets.MessagePacket:
						log.Info.Println("in manager message: MESSAGE", v)
						close(p.Receive)
						break
						updateOp :=  NewUpdateAndGetMembers(&user.User{Id:v.GetBy()},v.GetConvId())
						this.Pder.MakeOp() <- updateOp

						if r,ok := <-updateOp.Receive ; ok{
							p.Receive <- &InfoPacket{Paced:true,To:r}
						}
					case packets.EventPacket:
						log.Info.Println("in manager message: EVENT", v)
						close(p.Receive)
					}
				}
			}

		}

	}()
}

func Register(name string, provider Provider) {
	log.Warning.Printf("Register %s %$v", name, provider)
	if provider == nil {
		panic("message manager: Register provider is nil")
	}
	if _, dup := provides[name]; dup {
		panic("message manager: Register called twice for provider " + name)
	}
	provides[name] = provider
}
