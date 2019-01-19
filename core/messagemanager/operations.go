package messagemanager

import "github.com/Davidc2525/messager/user"

//interfas para operaciones
type Op interface {
	GetUser() *user.User
	GetInboxItemId() string
	Release()
}

//operacion de optener miembros de una conversacion
type GetMembers struct {
	User      *user.User
	InboxItem string
	Receive   chan []*user.User
}

func NewGetMembers(usr *user.User, inboxItem string) *GetMembers {
	return &GetMembers{User: usr, InboxItem: inboxItem, Receive: make(chan []*user.User)}
}

func (this *GetMembers) GetUser() *user.User {
	return this.User
}

func (this *GetMembers) Release() {
	close(this.Receive)
}

func (this *GetMembers) GetInboxItemId() string {
	return this.InboxItem
}

//operacion para actualizar el inbox
//la conversaion de los miembros de una conversacion
//y enviar los miembros
type UpdateAndGetMembers struct {
	User      *user.User
	InboxItem string
	Receive   chan []*user.User
}

func NewUpdateAndGetMembers(usr *user.User, inboxItem string) *UpdateAndGetMembers {
	return &UpdateAndGetMembers{User: usr, InboxItem: inboxItem, Receive: make(chan []*user.User)}
}

func (this *UpdateAndGetMembers) GetUser() *user.User {
	return this.User
}

func (this *UpdateAndGetMembers) Release() {
	close(this.Receive)
}

func (this *UpdateAndGetMembers) GetInboxItemId() string {
	return this.InboxItem
}
