package endpoint

import (
	"github.com/Davidc2525/messager/core/packets"
	"net/http"
)

type EndPointProvider interface {
	GetName() string
	SendTo() error
	Start() error
	Stop() error
}

//ConnectionEndPoint interface para conexiones de endpoints (puntos finales) q permite enviar y recibir
//paquetes de mensajes, eventos de los clientes (browser user)
type ConnectionEndPoint interface {
	Send() chan<- []byte
	//Receive() <-chan []byte
	OnClose() <-chan bool
	OnOpen() <-chan bool
	Close()

	IsOpen() bool

	Start()

	GetWriter() http.ResponseWriter
	GetRequest() *http.Request

	SetHandleClose(func())

	GetUid() string
	GetId() string

	Kind() packets.ClassConnection
	Host() string
}
