package message

import "encoding/json"

type Conpound struct {
	Plaint string `json:"plaint"`
	Draft  string `json:"draft"`
}

type CompoundContent struct {
	TType   uint8    `json:"t_type"`
	Content *Conpound `json:"content"`
}

func (this *CompoundContent) GetBytes() []byte {
	c, _ := json.Marshal(this.Content)
	return c
}

func NewCompoundContent() *CompoundContent {
	return &CompoundContent{TType: PLAINT_TEXT}
}

func (this *CompoundContent) GetType() uint8 {
	return this.TType
}

func (this *CompoundContent) GetContent() interface{} {
	return this.Content
}

func (this *CompoundContent) SetContent(content interface{}) {
	c:= Conpound{}
	json.Unmarshal(content.([]byte),c)
	this.Content = &c
}
