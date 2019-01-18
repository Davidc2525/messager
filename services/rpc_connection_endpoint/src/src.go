package src

var ()

type ArgsNewConnection struct {
	Cid    string
	Uid    string
	Hostid string
}

type ResponseNewConnection struct {
	Connected    bool
	RemoteHostId string
}

type ArgsNewConnectionPacket struct {
	Cid    string
	Uid    string
	Hostid string

	Packet []byte
}
