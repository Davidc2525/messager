/*
manager -> Provider -> StoreProvider
*/
package messagemanager

import (
	"github.com/Davidc2525/messager/core/messagemanager/inboxitem"
	"github.com/Davidc2525/messager/core/messagemanager/message"
	"github.com/Davidc2525/messager/core/messagemanager/storeprovider"
	"github.com/Davidc2525/messager/core/packets"
	"github.com/Davidc2525/messager/core/user"
	"github.com/Davidc2525/messager/log"
	"sync"
	"time"
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

/*
interface de alto nivel para mensajeria.
Se comunica con StoreProvider
*/
type Provider interface {
	//MakeOp() chan Op
	GetStore() storeprovider.StoreProvider

	GetMembersOfItem(usr *user.User, itemId string) ([]*user.User,error)
	GetMembersOfItemWithoutMe(usr *user.User, itemId string) ([]*user.User,error)
	UpdateAndGetMembersOfItem(usr *user.User, itemId string) ([]*user.User,error)
	UpdateAndGetItem(usr *user.User, itemId string) (*inboxitem.InboxItem,error)
	DeleteItem(usr *user.User, itemId string) error
}

type InfoPacket struct {
	Paced bool
	To    []*user.User
}

//paquete de proceso q contiene mensaje o evento producido,
//q tiene un canal para enviar informacion de vuelta
type ProcessPacket struct { //actor
	Receive chan *InfoPacket
	Packet  packets.Packet
}

/*
Administrador de mensajeria,
provee una interface basica para manejar los paquetes, tiene proveedor para
tener un acceso de mas bajo nivel:
manager -> Provider -> StoreProvider
*/
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
					log.Info.Printf("in manager : %#v\n", p)
					switch v := (p.Packet).(type) {
					case packets.MessagePacket:
						log.Info.Println("in manager message: MESSAGE", v)
						//close(p.Receive)
						//break
						//updateOp := NewUpdateAndGetMembers(&user.User{Id: v.GetBy()}, v.GetConvId())
						usr:=&user.User{Id: v.GetBy()}
						if item,err :=this.Pder.UpdateAndGetItem(usr,v.GetConvId()); err==nil{

							var id float64 = (1<<64-1) - float64(time.Now().UnixNano())
							msg := message.NewMessage(id)
							msg.Payload = v.GetTTypes()
							msg.By = v.GetBy()
							msg.SendAt = time.Now().Unix()*1000
							//msg.AppendTType(v.GetTType(),[]byte(v.GetMessage()))
							this.Pder.GetStore().StoreMessage(item,msg)

							members,_:=this.Pder.GetMembersOfItemWithoutMe(usr,v.GetConvId())
							p.Receive <- &InfoPacket{Paced: true, To: members}


						}else{
							p.Receive <- &InfoPacket{Paced: false}
						}

					case packets.EventPacket:
						log.Info.Println("in manager message: EVENT", v)
						if members,err :=this.Pder.GetMembersOfItemWithoutMe(&user.User{Id: v.GetBy()},v.GetConvId()); err==nil{
							p.Receive <- &InfoPacket{Paced: true, To: members}
						}else{
							p.Receive <- &InfoPacket{Paced: false}
						}
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
