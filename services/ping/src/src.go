package ping

type PingArgs struct {
	NodeId string
}

type PingResponse struct {
	Result bool
}

type PongArgs struct {
}

type PongResponse struct {
	Result bool
}
