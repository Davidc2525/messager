package inboxitem

import (
	"github.com/Davidc2525/messager/core/user"
	"sync"
	"time"
)

const (
	PRIVATE = uint8(iota)
	GROUP
)

type OpInboxItem interface {
	ReleaseOpInboxItem()
}
type GetMember struct {
	Receive chan *user.User
	Uid     string
}

func NewGetMember(uid string) *GetMember {
	getMember := &GetMember{Uid: uid, Receive: make(chan *user.User)}

	return getMember
}

func (this *GetMember) ReleaseOpInboxItem() {
	close(this.Receive)
}

type GetMembers struct {
	Receive chan []*user.User
}

func NewGetMembers() *GetMembers {
	return &GetMembers{Receive: make(chan []*user.User)}
}

func (this *GetMembers) ReleaseOpInboxItem() {
	close(this.Receive)
}

type RemoveMembers struct {
	Receive chan int // numero de miembro eliminados
	Users   []*user.User
}

func NewRemoveMembers(users []*user.User) *RemoveMembers {
	return &RemoveMembers{Users: users, Receive: make(chan int)}
}

func (this *RemoveMembers) ReleaseOpInboxItem() {
	close(this.Receive)
}

type AddMembers struct {
	Receive chan int
	Users   []*user.User
}

func NewAddMembers(users []*user.User) *AddMembers {
	return &AddMembers{Users: users, Receive: make(chan int)}
}

func (this *AddMembers) ReleaseOpInboxItem() {
	close(this.Receive)
}

type Member struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type InboxItem struct {
	//InOp      chan OpInboxItem      `json:"-"`
	Id        string                `json:"id"`
	Owner     string                `json:"owner"`
	Members   map[string]*user.User `json:"members"`
	CreatedAt int64                 `json:"created_at"`
	Type      uint8                   `json:"type"`
	lock      sync.Mutex
}

func NewInboxItem() *InboxItem {
	inboxItem := &InboxItem{Members: make(map[string]*user.User), CreatedAt: time.Now().Unix() * 1000}
	/*go func() {
		for {
			select {
			case op, ok := <-inboxItem.InOp:
				if ok {
					switch v := op.(type) {
					case *AddMembers:
						fmt.Printf("ADD MEMBER %#v\n", v)
						var adds int = 0
						for _, v := range v.Users {
							inboxItem.Members[v.Id] = v
							adds++
						}
						v.Receive <- adds
					case *GetMember:
						time.Sleep(time.Second * 2)
						if member, ok := inboxItem.Members[v.Uid]; ok {
							v.Receive <- member

						} else {
							close(v.Receive)
						}
					case *GetMembers:
						var members []*user.User
						for _, member := range inboxItem.Members {
							members = append(members, member)
						}
						v.Receive <- members
					case *RemoveMembers:
						var deletes int = 0
						for _, v := range v.Users {
							if _, ok := inboxItem.Members[v.Id]; ok {
								delete(inboxItem.Members, v.Id)
								deletes++
							}
						}
						v.Receive <- deletes
					}
				} else {
					return
				}
			}
		}
	}()*/
	return inboxItem
}

func (this *InboxItem) AddMember(m *user.User) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, exist := this.Members[m.Id]; !exist {
		this.Members[m.Id] = m
	}

}

func (this *InboxItem) RemoveMember(m *user.User) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, exist := this.Members[m.Id]; exist {
		delete(this.Members, m.Id)
	}
}

func (this *InboxItem) GetMembers() map[string]*user.User {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.Members
}

func (this *InboxItem) AddMembers(users []*user.User) int {
	this.lock.Lock()
	defer this.lock.Unlock()
	var adds int = 0
	for _, v := range users {
		this.Members[v.Id] = v
		adds++
	}
	return adds
}
