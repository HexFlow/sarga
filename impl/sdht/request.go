package sdht

type storeReq struct {
	ID   ID
	Key  string
	Data string
}

type findNodeReq struct {
	ID     ID
	FindID ID
}

type findNodeResp struct {
	Error error
	Peers []Peer
}

type findValueReq struct {
	ID  ID
	Key string
}

type findValueResp struct {
	Error error
	Data  []byte
}

type exitReq struct {
	ID ID
}
