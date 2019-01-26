package storememoryprovider

import (
	"fmt"
	"sync"
	"time"

	"github.com/Davidc2525/messager/core/messagemanager/inbox"
	"github.com/Davidc2525/messager/core/messagemanager/inboxitem"
	"github.com/Davidc2525/messager/core/messagemanager/message"
	"github.com/Davidc2525/messager/core/messagemanager/storeprovider"
	"github.com/Davidc2525/messager/core/user"
	mlog "github.com/Davidc2525/messager/log"
	"github.com/segmentio/ksuid"
)

var (
	log = mlog.New()
)

func makeKeyByMap(usrs map[string]*user.User) (r []*user.User) {
	adds := 0
	for _, u := range usrs {
		if adds >= 2 {
			break
		}
		r = append(r, u)
		adds++
	}
	return
}

func makeKey(usr1 *user.User, usr2 *user.User) string {
	return usr1.Id + ":" + usr2.Id
}

/*
Proveedor en memoria de inbox de usuario
*/
type MemoryStorage struct {
	/*
		Mapa de los inbox de usuario
		key -> set
		Uid : *inbox.Inbox
	*/
	Inboxs map[string]*inbox.Inbox
	/*
		Mantiene un mapa de los usuarios de las conversaciones q son privadas.
		eg:
			usuario 123 crea una conversacion privada con 55, se mapearia de la siguiente manera
			key -> set
			123:55 = itemId
			55:123 = itemId
	*/
	Privates map[string]string
	/*
		Sub-proveedor, un proveedor puede tener un sub proveedor para delegar funciones
		en este caso como este proveedor es en memoria, se le asigna un sub-proveedor de persistencia
	*/
	SubProvider storeprovider.StoreProvider

	lock sync.Mutex
	//InOp     chan storeprovider.StoreOp //borrar
}

func (this *MemoryStorage) StoreMessage(item *inboxitem.InboxItem, msg *message.Message) error {
	/*var network bytes.Buffer // Stand-in for the network.
	gob.NewEncoder(&network).Encode(msg)


	log.Warning.Printf("		STORE MESSAGE item: %#v\nmsg: %#v\nfrom gob: %#v", msg,network.String(),network)
	var m message.Message

	if err:= gob.NewDecoder(&network).Decode(&m);err !=nil{
		log.Error.Println(err.Error())
	}
	for _,t:= range m.Payload{
		log.Warning.Printf("		TTY id:%v, %v %#v",m.Id,t.Class, string(t.Data))
	}*/
	go this.GetSubProvider().StoreMessage(item, msg)
	return nil
}

func (this *MemoryStorage) MapPrivateItem(item *inboxitem.InboxItem) error {
	if item.Type == inboxitem.PRIVATE {
		uss := makeKeyByMap(item.GetMembers())

		key1 := makeKey(uss[0], uss[1])
		key2 := makeKey(uss[1], uss[0])

		this.Privates[key1] = item.Id
		this.Privates[key2] = item.Id

		this.GetSubProvider().MapPrivateItem(item)

		return nil

	} else {
		//no es privado
		return fmt.Errorf("el item %v no es privado", item.Id)
	}
}

func (this *MemoryStorage) UnMapPrivateItem(item *inboxitem.InboxItem) error {
	if item.Type == inboxitem.PRIVATE {
		uss := makeKeyByMap(item.GetMembers())

		key1 := makeKey(uss[0], uss[1])
		key2 := makeKey(uss[1], uss[0])

		delete(this.Privates, key1)
		delete(this.Privates, key2)
		this.GetSubProvider().UnMapPrivateItem(item)

		return nil

	} else {
		//no es privado
		return fmt.Errorf("el item %v no es privado", item.Id)
	}
}

func (this *MemoryStorage) GetPrivateItem(usr1 *user.User, usr2 *user.User) (string, error) {

	key1 := makeKey(usr1, usr2)
	key2 := makeKey(usr2, usr1)

	return this.GetSubProvider().GetPrivateItem(usr1, usr2)
	if idItem, ok := this.Privates[key1]; ok {
		return idItem, nil
	} else if idItem, ok := this.Privates[key2]; ok {
		return idItem, nil
	} else if s := this.GetSubProvider(); s != nil {
		if idItem, err := this.GetSubProvider().GetPrivateItem(usr1, usr2); err == nil {
			return idItem, err
		} else {
			return idItem, err
		}
	} else {
		return "", fmt.Errorf("no se econtro ningun item")
	}
}

func (this *MemoryStorage) GetInbox(usr *user.User) (*inbox.Inbox, error) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if in, ok := this.Inboxs[usr.Id]; ok {
		return in, nil
	} else {
		if this.GetSubProvider() != nil {
			if in, err := this.GetSubProvider().GetInbox(usr); err == nil {
				this.Inboxs[usr.Id] = in
				return in, nil
			} else {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("el usuario %v no tiene inbox", usr.Id)
		}

	}
	return nil, nil
}

func (this *MemoryStorage) CreateInboxWith(usr *user.User, in *inbox.Inbox) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.Inboxs[usr.Id]; ok {
		return fmt.Errorf("ya existe inbox para %#v", usr)
	} else {
		//this.Inboxs[usr.Id] = in
		if this.GetSubProvider() != nil {
			if err := this.GetSubProvider().CreateInboxWith(usr, in); err != nil {
				log.Warning.Println("no se pudo crear el inbox en sub provider: ", err)
			} else {
				this.Inboxs[usr.Id] = in
			}
		} else {
			this.Inboxs[usr.Id] = in
		}
		return nil
	}
	return nil
}

func (this *MemoryStorage) SaveInbox(usr *user.User, in *inbox.Inbox) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if this.GetSubProvider() != nil {
		return this.SubProvider.SaveInbox(usr, in)
	}
	return fmt.Errorf("no provee de un sub proveedor de persistencia")
}

func (this *MemoryStorage) CreateInbox(usr *user.User, members []*user.User) (*inbox.Inbox, error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	if _, ok := this.Inboxs[usr.Id]; !ok {
		id := ksuid.New().String()
		newInbox := inbox.NewInbox()
		newInbox.Id = id
		newInbox.Owner = usr

		if this.GetSubProvider() != nil {
			if err := this.GetSubProvider().CreateInboxWith(usr, newInbox); err != nil {
				log.Warning.Println("no se pudo crear el inbox en sub provider: ", err)
			} else {
				this.Inboxs[usr.Id] = newInbox
			}
		} else {
			this.Inboxs[usr.Id] = newInbox
		}
		return newInbox, nil
	} else {

	}
	return nil, nil
}

func (this *MemoryStorage) DeleteInbox(usr *user.User, in *inbox.Inbox) error {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.Inboxs[usr.Id]; ok {
		delete(this.Inboxs, usr.Id)
		if this.GetSubProvider() != nil { //buskar en sub provider si lo tiene,
			this.GetSubProvider().DeleteInbox(usr, in)
			return nil
		} else {
			return fmt.Errorf("no existe inbox para eliminar")
		}
	} else {
		return fmt.Errorf("no existe inbox para eliminar")
	}

	return nil
}

//guardar todos los inbox en memoria
func (this *MemoryStorage) SaveAllInbox() error {
	//this.lock.Lock()
	//defer this.lock.Unlock()
	for k, v := range this.Inboxs {
		log.Warning.Printf("Guardando inbox de %v", k)
		this.SaveInbox(&user.User{Id: k}, v)
	}
	return nil
}

func (this *MemoryStorage) SetSubProvider(sub storeprovider.StoreProvider) {
	this.SubProvider = sub
}

func (this *MemoryStorage) GetSubProvider() storeprovider.StoreProvider {
	return this.SubProvider
}

func NewMemoryStorage() *MemoryStorage {
	t := time.NewTicker(time.Second * 5)
	sp := &MemoryStorage{Privates: make(map[string]string), Inboxs: make(map[string]*inbox.Inbox)}
	go func() {
		for {
			select {
			case <-t.C:
				log.Warning.Printf("Guardando todos los inbox...")
				sp.SaveAllInbox()
			}
		}

	}()
	return sp
}
