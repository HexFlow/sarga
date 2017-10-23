package apiserver

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net/http"

	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
)

func StartAPIServer(args ServerArgs, dht dht.DHT) {
	addr := iface.GetAddress(args.IP, args.Port, args.Proto)
	s := &http.Server{
		Addr:    addr.ToString(),
		Handler: &proxyHandler{dht},
	}
	err := s.ListenAndServe()
	if err != nil {
		log.Println(err)
	}
	dht.Shutdown()
}

type proxyHandler struct {
	dht dht.DHT
}

func (h *proxyHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	path := req.URL.String()
	body := []byte{}
	_, err := req.Body.Read(body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
	}

	if req.Method == "GET" {
		hash := sha1.New()
		io.WriteString(hash, req.URL.RequestURI())
		// TODO: Assert 160 bytes length
		data, err := h.dht.FindValue(hex.EncodeToString(hash.Sum(nil)))
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			_, err := rw.Write([]byte(err.Error()))
		}
	} else if req.Method == "POST" {
		data := []byte{}
		_, err := req.Body.Read(data)
		if err == nil {
			hash := sha1.New()
			io.WriteString(hash, req.URL.RequestURI())
			// TODO: Assert 160 bytes length
			err = h.dht.StoreValue(hex.EncodeToString(hash.Sum(nil)), data)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				_, err := rw.Write([]byte(err.Error()))
			}
		}
	} else {
		rw.WriteHeader(http.StatusBadRequest)
		_, err = rw.Write([]byte("Unsupported method. Allowed methods: GET, POST"))
	}

	if err != nil {
		log.Println(err)
	}
}
