package message

const (
	PLAINT_TEXT     = uint8(iota) //representacion pana del mensaje
	DRAFT_OBJECT                  //contiene representacion en formato draftjs
	FILE_POINTER                  //apunta a un id q contiene informacion de un archivo //TODO
	COMPOUND_OBJECT               //contiene representacion plana del mensaje junto con la de draftjs
)

type TType struct {
	Class uint8  `json:"class"`
	Data  string `json:"data"`
}

type Message struct {
	By      string   `json:"by"`
	Id      float64  `json:"id"`
	SendAt  int64    `json:"send_at"`
	Payload []*TType `json:"payload"`
}

func NewMessage(id float64) *Message {
	return &Message{Id: id}
}

func (this *Message) AppendTType(class uint8, data string) {
	this.Payload = append(this.Payload, &TType{Class: class, Data: data})
}
