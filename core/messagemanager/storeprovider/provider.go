package storeprovider

import (
	"github.com/Davidc2525/messager/core/messagemanager/inbox"
	"github.com/Davidc2525/messager/user"
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

type StoreProvider interface {
	MakeOp() chan StoreOp

	/*GetInbox(user *user.User) *inbox.Inbox
	SaveInbox(user *user.User, inbox *inbox.Inbox)
	CreateInbox(user *user.User, members []*user.User) *inbox.Inbox
	DeleteInbox(user *user.User, inbox *inbox.Inbox) error

	SaveAllInbox() error*/
}
