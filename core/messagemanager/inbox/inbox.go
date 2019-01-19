package inbox

import (
	"fmt"
	. "github.com/Davidc2525/messager/core/messagemanager/inboxitem"
	"sync"
)

const (
	FROMT = iota
	BACK
	IN
)

type Op interface {
}

type Put struct {
	Op
	Input  *InboxItem
	Method int
	Index  int
}

type Get struct {
	Op
	Receive chan *GetResult
	Id      string
}

func NewGet() *Get {
	return &Get{Receive: make(chan *GetResult)}
}

type GetResult struct {
	Item  *InboxItem
	Index int
}

func (this *Get) release() {
	close(this.Receive)
}

type Move struct {
	Op
	Input  *InboxItem
	Method int
	Index  int
}

type Del struct {
	Op
	Input *InboxItem
	Index int
}

type Inbox struct {
	Id    string       `json:"id"`
	Items []*InboxItem `json:"items"`
	lock  sync.Mutex   `json:"-"`

	Close chan bool `json:"-"`

	Op chan Op `json:"-"`
}

func NewInbox() *Inbox {
	inbox := &Inbox{Items: []*InboxItem{}, Close: make(chan bool), Op: make(chan Op, 10)}
	go func() {

		for {
			select {
			case <-inbox.Close:
				close(inbox.Close)
				close(inbox.Op)
				inbox.Items = nil
				return
			case op, _ := <-inbox.Op:

				switch v := op.(type) {
				case Put:
					switch v.Method {
					case BACK:
						inbox.PutBack(v.Input)
					case FROMT:
						inbox.PutFromt(v.Input)
					case IN:
						inbox.PutIn(v.Input, v.Index)
					}
				case Del:
					if v.Input != nil {
						inbox.DeleteIn(inbox.IndexOf(v.Input))
					} else {
						inbox.DeleteIn(v.Index)
					}

				case Move:
					switch v.Method {
					case BACK:
						inbox.MoveBack(v.Input)
					case FROMT:
						inbox.MoveFromt(v.Input)
					case IN:
						inbox.MoveTo(v.Input, v.Index)
					}

				case Get:
					index := inbox.IndexOfById(v.Id)
					item, err := inbox.Get(index)
					if err == nil {
						v.Receive <- &GetResult{item, index}
					} else {
						close(v.Receive)
					}

				}
			}
		}

	}()
	return inbox
}

func (this *Inbox) PutFromt(item *InboxItem) {
	this.PutIn(item, 0)
}

func (this *Inbox) PutBack(item *InboxItem) {
	//this.PutIn(item, len(this.Items))
	this.Items = append(this.Items, item)
}
func (this *Inbox) PutIn(item *InboxItem, index int) {
	//this.lock.Lock()
	//defer this.lock.Unlock()
	if index > len(this.Items) {
		index = len(this.Items)
	}
	if index < 0 {
		index = 0
	}
	e := []*InboxItem{}
	e = append(e, this.Items[:index]...)
	e = append(e, item)
	e = append(e, this.Items[index:]...)
	this.Items = e
}
func (this *Inbox) DeleteIn(index int) {
	//this.lock.Lock()
	//defer this.lock.Unlock()
	if index < 0 {
		return
	}

	index2 := index + 1

	if index2 > len(this.Items) {
		index2 = len(this.Items)
	}

	e := append([]*InboxItem{}, this.Items[:index]...)
	e = append(e, this.Items[index2:]...)
	this.Items = e
}

func (this *Inbox) MoveFromt(item *InboxItem) {
	index := this.IndexOf(item)
	if index < 0 {
		return
	}
	this.DeleteIn(this.IndexOf(item))
	this.PutFromt(item)
}

func (this *Inbox) MoveBack(item *InboxItem) {
	index := this.IndexOf(item)
	if index < 0 {
		return
	}
	this.DeleteIn(index)
	this.PutBack(item)
}
func (this *Inbox) MoveTo(item *InboxItem, index int) {
	indexItem := this.IndexOf(item)
	if indexItem < 0 {
		return
	}
	this.DeleteIn(indexItem)
	this.PutIn(item, index)
}
func (this *Inbox) Get(index int) (*InboxItem, error) {
	//this.lock.Lock()
	//defer this.lock.Unlock()
	if index > len(this.Items)-1 || index < 0 {
		return nil, fmt.Errorf("Item %v no exist", index)
	}
	return this.Items[index], nil
}

func (this *Inbox) IndexOf(item *InboxItem) int {
	//this.lock.Lock()
	//defer this.lock.Unlock()
	var index = -1
	for k, i := range this.Items {
		if i == item {
			index = k
			break
		}
	}
	return index
}

func (this *Inbox) IndexOfById(item string) int {
	//this.lock.Lock()
	//defer this.lock.Unlock()
	var index = -1
	for k, i := range this.Items {
		if i.Id == item {
			index = k
		}
	}
	return index
}
