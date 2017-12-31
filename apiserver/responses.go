package apiserver

type getAttrResp struct {
	FileType FileType
}

type readDirResp struct {
	Files []string
}

type readResp []byte
