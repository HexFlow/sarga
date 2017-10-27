package apiserver

import (
	"log"
	"net/http"
	"strings"

	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
)

func StartAPIServer(args ServerArgs, dht dht.DHT) {
	addr := iface.GetAddress(args.IP, args.Port)
	s := &http.Server{
		Addr:    addr.ToString(),
		Handler: &proxyHandler{dht},
	}
	log.Println("Listening on", addr.ToString())
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
	body := []byte{}
	_, err := req.Body.Read(body)
	if err != nil {
		rw.WriteHeader(http.StatusBadRequest)
	}

	if req.URL.Hostname() == "sarga" {
		// Admin panel / upload UI / statistics visualization is done on http://sarga domain.

		if strings.HasPrefix(req.URL.RequestURI(), "/upload/") {
			// Upload file.
			path := string(req.URL.RequestURI()[7]) // drop the initial /upload part.
			data := []byte{}
			_, err := req.Body.Read(data)
			if err == nil {
				err = uploadFile(path, data, h.dht)
				if err != nil {
					rw.WriteHeader(http.StatusInternalServerError)
					rw.Write([]byte(err.Error()))
				}
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte("File uploaded at " + path))
			}
		}
	} else {
		// All non-sarga domain proxy server requests will be treated as fetch requests.
		// TODO(sakshams): Filter requests on the hostname. If the hostname is not
		// http://sarga-fetch domain, let the request pass through to the regular internet.
		if req.Method == "GET" {
			// Fetch file.
			path := string(req.URL.RequestURI())
			data, err := downloadFile(path, h.dht)
			if err != nil {
				rw.WriteHeader(http.StatusNotFound)
				rw.Write([]byte(err.Error()))
			}
			rw.WriteHeader(http.StatusOK)
			rw.Write(data)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			_, err = rw.Write([]byte("Unsupported method. Allowed methods: GET"))
		}
		if err != nil {
			log.Println(err)
		}
	}

}
