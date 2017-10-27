package sdht

type storeReq struct {
	ID   ID
	Key  string
	Data string
}

type findNodeReq struct {
	ID  ID
	Key string
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
	Peers []Peer
}

type exitReq struct {
	ID ID
}

type pingResp struct {
	ID ID
}

type fakeHTTP struct {
	path string
	data []byte
}
