package message

type PlaintTextContent struct {
	TType   uint8  `json:"t_type"`
	Content string `json:"content"`
}

func NewPlaintTextContent() *PlaintTextContent {
	return &PlaintTextContent{TType: PLAINT_TEXT}
}

func (this *PlaintTextContent) GetBytes() []byte {
	return []byte(this.Content)
}

func (this *PlaintTextContent) GetType() uint8 {
	return this.TType
}

func (this *PlaintTextContent) GetContent() interface{} {
	return this.Content
}

func (this *PlaintTextContent) SetContent(content interface{}) {
	this.Content = content.(string)
}
