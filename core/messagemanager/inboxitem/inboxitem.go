package inboxitem

import (
	"github.com/Davidc2525/messager/user"
	"sync"
	"time"
)

type Member struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type InboxItem struct {
	Id        string    `json:"id"`
	Owner     string    `json:"owner"`
	Members   []*user.User `json:"members"`
	CreatedAt int64     `json:"created_at"`
	lock      sync.Mutex
}

func NewInboxItem() *InboxItem {
	return &InboxItem{Members: []*user.User{}, CreatedAt: time.Now().UnixNano()}
}

func (this *InboxItem) AddMember(m *user.User) {
	defer this.lock.Unlock()

	if this.IndexOfMemberId(m.Id) == -1 {
		this.lock.Lock()
		this.Members = append(this.Members, m)
	}
}

func (this *InboxItem) RemoveMember(m *user.User) {
	index := this.IndexOfMember(m)
	this.lock.Lock()
	defer this.lock.Unlock()
	if index < 0 {
		return
	}

	index2 := index + 1

	if index2 > len(this.Members) {
		index2 = len(this.Members)
	}

	e := append([]*user.User{}, this.Members[:index]...)
	e = append(e, this.Members[index2:]...)
	this.Members = e
}

func (this *InboxItem) IndexOfMember(m *user.User) int {
	this.lock.Lock()
	defer this.lock.Unlock()
	var index int
	index = -1
	for k, memberInItem := range this.Members {
		if m == memberInItem {
			index = k
			break
		}
	}
	return index
}

func (this *InboxItem) IndexOfMemberId(m string) int {
	this.lock.Lock()
	defer this.lock.Unlock()
	var index int
	index = -1
	for k, memberInItem := range this.Members {
		if m == memberInItem.Id {
			index = k
			break
		}
	}
	return index
}
