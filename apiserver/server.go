package apiserver

import (
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
)

// TODO(sakshams): Should have a shutdown channel for integration tests.
func StartAPIServer(args iface.CommonArgs, dht dht.DHT) {
	addr := iface.GetAddress(args.IP, args.Port)
	s := &http.Server{
		Addr:    addr.String(),
		Handler: &proxyHandler{dht},
	}
	log.Println("Listening on", addr.String())
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
	path := req.URL.RequestURI()
	if strings.HasPrefix(path, "/sarga/") {
		path = path[6:] // Drop the initial /sarga part.

		if strings.HasPrefix(path, "/upload/") {
			// Upload file.
			path := path[7:] // Drop the initial /upload part.
			data, err := ioutil.ReadAll(req.Body)
			if err == nil {
				err = uploadFile(path, data, h.dht)
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write([]byte(err.Error()))
				} else {
					rw.WriteHeader(http.StatusOK)
					rw.Write([]byte("File uploaded at " + path))
				}
			}

		} else {
			if req.Method == "GET" {
				// Fetch file.
				data, err := downloadFile(path, h.dht)
				if err != nil {
					rw.WriteHeader(http.StatusNotFound)
					rw.Write([]byte(err.Error()))
					log.Println(err)
				} else {
					rw.WriteHeader(http.StatusOK)
					rw.Write(data)
				}
			} else {
				rw.WriteHeader(http.StatusBadRequest)
				_, err := rw.Write([]byte("Unsupported method. Allowed methods: GET"))
				if err != nil {
					log.Println(err)
				}
			}
		}
	} else {
		// TODO(sakshams): Forward regular requests to the internet.
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte("Forwarding non-/sarga/ requests is not supported yet."))
		if err != nil {
			log.Println(err)
		}
	}
}
