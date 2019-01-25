package packets

import (
	"github.com/Davidc2525/messager/core/messagemanager/message"
	"github.com/Davidc2525/messager/core/user"
	"github.com/gorilla/websocket"
)

type ClassConnection int

const (
	Local  ClassConnection = 0
	Remote ClassConnection = 1
)

type Container interface {
	GetClass() ClassConnection
	GetHost() string
	IsForWard() bool
	GetData() []byte
	GetCid() string
}

type DContainer struct {
	Cid     string          `json:"cid"`
	Class   ClassConnection `json:"class"`
	Host    string          `json:"host"`
	Forward bool            `json:"forward"`
	Data    []byte          `json:"data"`
}

func (this *DContainer) GetCid() string {
	return this.Cid
}

func NewDContainer(class ClassConnection, host string, forward bool, data []byte) *DContainer {
	return &DContainer{Class: class, Host: host, Forward: forward, Data: data}
}

func (this *DContainer) GetClass() ClassConnection {
	return this.Class
}

func (this *DContainer) GetHost() string {
	return this.Host
}

func (this *DContainer) IsForWard() bool {
	return this.Forward
}

func (this *DContainer) GetData() []byte {
	return this.Data
}

/*interfaces*/

//packet parent
type Packet interface {
	GetKind() string

	SetAttr(key string, value string)
	DelAttr(key string)
	GetAttr(key string) string
}

type EventPacket interface {
	Packet
	GetEvent() string
	GetBy() string
	GetTo() string
	GetId() string
	GetConvId() string
}

type MessagePacket interface {
	Packet
	GetMessage() string
	GetTType() uint8
	GetBy() string
	GetTo() string
	GetConvId() string
	GetTTypes()[]*message.TType
}

type PrecensePacket interface {
	Packet
}

type UserPacket interface {
	Packet
	User() *user.User
}

/*Implementaciones*/

type DEventPacket struct {
	Kind           string            `json:"kind"`
	Attr           map[string]string `json:"attr"`
	IdConversation string            `json:"id_conversation"`
	Event          string            `json:"event"`
	By             string            `json:"by"`
	To             string            `json:"to"`
}

func (this *DEventPacket) GetConvId() string {
	return this.IdConversation
}

func NewDEventPacket() *DEventPacket {
	return &DEventPacket{Kind: "event", Attr: make(map[string]string)}
}

func (this *DEventPacket) GetKind() string {
	return this.Kind
}

func (this *DEventPacket) SetAttr(key string, value string) {
	this.Attr[key] = value
}

func (this *DEventPacket) DelAttr(key string) {
	delete(this.Attr, key)
}

func (this *DEventPacket) GetAttr(key string) string {
	return this.Attr[key]
}

func (this *DEventPacket) GetEvent() string {
	return this.Event
}
func (this *DEventPacket) GetTo() string {
	return this.To
}

func (this *DEventPacket) GetBy() string {
	return this.By
}

func (this *DEventPacket) GetId() string {
	return this.IdConversation
}

//MessagePacket implementation
type DMessagePacket struct {
	TType          uint8             `json:"t_type"`
	Attr           map[string]string `json:"attr"`
	Kind           string            `json:"kind"`
	IdConversation string            `json:"id_conversation"`
	By             string            `json:"by"`
	To             string            `json:"to"`
	Message        string            `json:"message"`
	Payload        []*message.TType  `json:"payload"`
}
func (this *DMessagePacket) GetTTypes() []*message.TType{
	return this.Payload
}

func (this *DMessagePacket) SetAttr(key string, value string) {
	this.Attr[key] = value
}

func (this *DMessagePacket) DelAttr(key string) {
	delete(this.Attr, key)
}

func (this *DMessagePacket) GetAttr(key string) string {
	if v, ok := this.Attr[key]; ok {
		return v
	}
	return ""
}

func (this *DMessagePacket) GetBy() string {
	return this.By
}

func (this *DMessagePacket) GetTo() string {
	return this.To
}

func (this *DMessagePacket) GetConvId() string {
	return this.IdConversation
}

func NewDMessagePacket() *DMessagePacket {
	return &DMessagePacket{Kind: "message", Attr: make(map[string]string),Payload:[]*message.TType{}}
}

func (this *DMessagePacket) GetKind() string {
	return this.Kind
}

func (this *DMessagePacket) GetMessage() string {
	return this.Message
}

func (this *DMessagePacket) GetTType() uint8 {
	return this.TType
}

/*func (this *DMessagePacket) GetContent() message.Content {
	var c message.Content
	switch this.TypeContent {
	case message.PLAINT_TEXT:
		c := message.NewPlaintTextContent()
		c.SetContent(this.Message)

		return c
	}
	return c
}*/

//UserPacket implementation
type DUserPacket struct {
	Kind string
	Us   *user.User
	Conn *websocket.Conn
	Id   string
}

func (this *DUserPacket) SetAttr(key string, value string) {
	panic("implement me")
}

func (this *DUserPacket) DelAttr(key string) {
	panic("implement me")
}

func (this *DUserPacket) GetAttr(key string) string {
	panic("implement me")
}

func (this *DUserPacket) User() *user.User {
	return this.Us
}

func NewDUserPacket() *DUserPacket {
	return &DUserPacket{Kind: "user"}
}

func (this *DUserPacket) GetKind() string {
	return this.Kind
}
