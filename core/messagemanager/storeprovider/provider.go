package storeprovider

import (
	"github.com/Davidc2525/messager/core/messagemanager/inbox"
	"github.com/Davidc2525/messager/core/messagemanager/inboxitem"
	"github.com/Davidc2525/messager/core/messagemanager/message"
	"github.com/Davidc2525/messager/core/user"
)

type StoreOp interface {
	ReleaseOp() //metodo solo para implementar
}
type GetInbox struct {
	User    *user.User
	Receive chan *inbox.Inbox
}

func NewGetInbox(user *user.User) *GetInbox {
	return &GetInbox{User: user, Receive: make(chan *inbox.Inbox)}
}

func (this *GetInbox) ReleaseOp() {
	close(this.Receive)
}

type SaveInbox struct {
	User    *user.User
	Receive chan *inbox.Inbox
}

func NewSaveInbox(user *user.User) *SaveInbox {
	return &SaveInbox{User: user, Receive: make(chan *inbox.Inbox)}
}

func (this *SaveInbox) ReleaseOp() {
	close(this.Receive)
}

type CreateInbox struct {
	User    *user.User
	Members []*user.User
	Receive chan *inbox.Inbox
}

func NewCreateInbox(user *user.User, members []*user.User) *CreateInbox {
	return &CreateInbox{User: user, Members: members, Receive: make(chan *inbox.Inbox)}
}

func (this *CreateInbox) ReleaseOp() {
	close(this.Receive)
}

type CreateInboxWith struct {
	User    *user.User
	Inbox   *inbox.Inbox
	Members []*user.User
	Receive chan *inbox.Inbox
}

func NewCreateInboxWith(user *user.User, in *inbox.Inbox) *CreateInboxWith {
	return &CreateInboxWith{User: user, Inbox: in, Receive: make(chan *inbox.Inbox)}
}

func (this *CreateInboxWith) ReleaseOp() {
	close(this.Receive)
}

type DeleteInbox struct {
	User    *user.User
	Receive chan bool
}

func NewDeleteInbox(user *user.User) *DeleteInbox {
	return &DeleteInbox{User: user, Receive: make(chan bool)}
}

func (this *DeleteInbox) ReleaseOp() {
	close(this.Receive)
}

type SaveAllInbox struct {
	Receive chan bool
}

func NewSaveAllInbox() *SaveAllInbox {
	return &SaveAllInbox{}
}

func (this *SaveAllInbox) ReleaseOp() {
	close(this.Receive)
}

/*
Proveedor de almacenamiento de inbox, para proveedor de mensajeria.
stack: StoreProvider -> Provider -> Manager
*/
type StoreProvider interface {
	//MakeOp() chan StoreOp
	GetSubProvider() StoreProvider
	SetSubProvider(sub StoreProvider)

	GetInbox(usr *user.User) (*inbox.Inbox, error)
	SaveInbox(usr *user.User, inbox *inbox.Inbox) error
	CreateInbox(usr *user.User, members []*user.User) (*inbox.Inbox, error)
	CreateInboxWith(usr *user.User, in *inbox.Inbox) error
	DeleteInbox(usr *user.User, inbox *inbox.Inbox) error

	SaveAllInbox() error

	MapPrivateItem(item *inboxitem.InboxItem) error
	UnMapPrivateItem(item *inboxitem.InboxItem) error
	GetPrivateItem(usr1 *user.User, usr2 *user.User) (string, error)

	//messages

	StoreMessage(item *inboxitem.InboxItem, msg *message.Message) error
}
