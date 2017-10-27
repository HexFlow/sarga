package httpnet

import (
	"bytes"
	"net/http"

	"github.com/sakshamsharma/sarga/common/iface"
)

// HTTPNet is a minimal implementation of a HTTPNetwork (network.HTTPNetwork) to be used
// with sarga. It uses HTTP as the middleware protocol.
type HTTPNet struct {
}

var _ iface.Net = &HTTPNet{}

func (n *HTTPNet) Get(addr iface.Address, path string) ([]byte, error) {
	resp, err := http.Get(addr.String())
	if err != nil {
		return nil, err
	}
	// TODO
	bytes := []byte{}
	_, _ = resp.Body.Read(bytes)
	return bytes, nil
}

func (n *HTTPNet) Put(addr iface.Address, path string, data []byte) error {
	_, err := http.Post(addr.String()+"/"+path, "application/json", bytes.NewBuffer(data))
	return err
}

func (n *HTTPNet) Post(addr iface.Address, path string, data []byte) ([]byte, error) {
	resp, err := http.Post(addr.String()+"/"+path, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	// TODO
	bytes := []byte{}
	_, _ = resp.Body.Read(bytes)
	return bytes, nil
}

func (n *HTTPNet) Listen(addr iface.Address, handler func(string, []byte) []byte, shutdown chan bool) error {
	s := &http.Server{
		Addr:    addr.String(),
		Handler: &httphandler{handler},
	}
	go s.ListenAndServe()
	for {
		select {
		case <-shutdown:
			s.Close()
			return nil
		}
	}
}

type httphandler struct {
	handler func(string, []byte) []byte
}

func (h *httphandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.String()
	body := []byte{}
	// TODO: Catch error
	_, _ = req.Body.Read(body)
	// TODO: Catch error
	_, _ = rw.Write(h.handler(path, body))
}
