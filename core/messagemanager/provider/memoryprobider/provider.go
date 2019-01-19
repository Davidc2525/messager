package memoryprobider

import (
	"github.com/Davidc2525/messager/core/messagemanager"
	"github.com/Davidc2525/messager/core/messagemanager/inbox"
	"github.com/Davidc2525/messager/core/messagemanager/storeprovider"
)

type MemoryStorage struct {
	Inboxs map[string]*inbox.Inbox
	InOp   chan storeprovider.StoreOp
}

func (this *MemoryStorage) MakeOp() chan storeprovider.StoreOp {
	return this.InOp
}

func NewMemoryStorage() *MemoryStorage {
	sp := &MemoryStorage{}
	go func() {
		for {
			select {
			case op := <-sp.InOp:
				switch v := op.(type) {
				case *storeprovider.GetInbox:
					if inb, ok := sp.Inboxs[v.User.Id]; ok {
						v.Receive <- inb
					} else { //buskar en db
						v.ReleaseOp()
					}
				}
			}
		}
	}()
	return sp
}

type MemoryProvider struct {
	Op chan messagemanager.Op

	StoreProvider storeprovider.StoreProvider
}

func (this *MemoryProvider) Start() {
	go func() {
		for {
			select {
			case op := <-this.Op:
				switch v := op.(type) {
				case *messagemanager.GetMembers:

				case *messagemanager.UpdateAndGetMembers:
					getInboxOp := storeprovider.NewGetInbox(v.GetUser())

					this.StoreProvider.MakeOp() <- getInboxOp

					if r, ok := <-getInboxOp.Receive; ok { //exist

						getInboxOp.ReleaseOp()

						iigetop := inbox.NewGet()
						iigetop.Id = v.GetInboxItemId()

						r.Op <- iigetop

						if reg, ok := <-iigetop.Receive; ok {

							r.Op <- inbox.Move{Method: inbox.FROMT, Input: reg.Item}

							v.Receive <- reg.Item.Members
						}
					} else { //no exist inbox, search in db

					}

				}
			}
		}
	}()
}

func NewMemoryProvider() *MemoryProvider {
	return &MemoryProvider{StoreProvider: &MemoryStorage{Inboxs: make(map[string]*inbox.Inbox)}}
}
func init() {
	//messagemanager.Register("memory",NewMemoryProvider())
}

func (this *MemoryProvider) MakeOp() chan messagemanager.Op {
	return this.Op
}
