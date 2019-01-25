package message

type Content interface {
	GetType() uint8
	//SetType () uint8
	GetContent() interface{}
	GetBytes() []byte
	SetContent(content interface{})
}
