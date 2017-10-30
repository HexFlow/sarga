package httpnet

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/sakshamsharma/sarga/common/iface"
)

// HTTPNet is a minimal implementation of a HTTPNetwork (network.HTTPNetwork) to be used
// with sarga. It uses HTTP as the middleware protocol.
type HTTPNet struct {
}

var _ iface.Net = &HTTPNet{}

func (n *HTTPNet) Get(addr iface.Address, path string) ([]byte, error) {
	resp, err := http.Get("http://" + addr.String() + "/" + path)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (n *HTTPNet) Put(addr iface.Address, path string, data []byte) error {
	bufReader := ioutil.NopCloser(bytes.NewBuffer(data))
	_, err := http.Post("http://"+addr.String()+"/"+path, "text/plain", bufReader)
	return err
}

func (n *HTTPNet) Post(addr iface.Address, path string, data []byte) ([]byte, error) {
	bufReader := ioutil.NopCloser(bytes.NewBuffer(data))
	resp, err := http.Post("http://"+addr.String()+"/"+path, "text/plain", bufReader)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (n *HTTPNet) Listen(addr iface.Address, handler func(string, []byte) []byte, shutdown chan bool) error {
	s := &http.Server{
		Addr:    addr.String(),
		Handler: &httphandler{handler},
	}
	go func() {
		err := s.ListenAndServe()
		if err != nil {
			log.Println("HTTPNet listen threw an error:", err)
		} else {
			log.Println("HTTPNet listen terminated")
		}
	}()
	fmt.Println("HTTPNet Listening on", addr)
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
	rw.Header().Set("Access-Control-Allow-Origin", "*")
	rw.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	rw.Header().Set("Access-Control-Allow-Headers",
		"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	// Stop here if its Preflighted OPTIONS request
	if req.Method == "OPTIONS" {
		return
	}

	path := strings.TrimPrefix(req.URL.Path, "/")
	defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
		rw.Write([]byte("Could not read request body"))
	} else {
		// TODO: Catch error
		_, _ = rw.Write(h.handler(path, body))
	}
}
