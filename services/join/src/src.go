package join

type JoinArgs struct {
	NodeId string
}

type JoinResponse struct {
	Result       bool
	RemoteHostId string
}
