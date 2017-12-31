package apiserver

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/elazarl/goproxy"

	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
)

type handleFuncType func(rw http.ResponseWriter, req *http.Request)

// TODO(sakshams): Should have a shutdown channel for integration tests.
func StartAPIServer(args iface.CommonArgs, dht dht.DHT) {
	h := &proxyHandler{dht}
	fs := http.FileServer(http.Dir("static"))

	http.HandleFunc("/sarga/upload/", prefixHandler("/sarga/upload", h.uploadHandler))
	http.HandleFunc("/sarga/files/", prefixHandler("/sarga/files", h.filesHandler))
	http.HandleFunc("/sarga/info/", prefixHandler("/sarga/info", h.apiHandler))
	http.HandleFunc("/sarga/fs/", prefixHandler("/sarga/fs", h.fsHandler))
	http.Handle("/sarga/", http.StripPrefix("/sarga", fs))
	http.Handle("/", goproxy.NewProxyHttpServer())

	addr := iface.GetAddress(args.IP, args.Port).String()
	log.Println("Listening on", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Println(err)
	}
	dht.Shutdown()
}

// prefixHandler creates a handler which receives a stripped req.URL.Path.
func prefixHandler(prefix string, handler handleFuncType) handleFuncType {
	return func(rw http.ResponseWriter, req *http.Request) {
		req.URL.Path = req.URL.Path[len(prefix):]
		handler(rw, req)
	}
}

type proxyHandler struct {
	dht dht.DHT
}

func (h *proxyHandler) fsHandler(rw http.ResponseWriter, req *http.Request) {
	url := req.URL.Path

	if strings.HasPrefix(url, "/getattr") {
		url = url[len("/getattr"):]
		data, err := fsGetAttr(url, h.dht)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			log.Println(err)
		} else {
			rw.WriteHeader(http.StatusOK)
			rw.Write(data)
		}
	} else if strings.HasPrefix(url, "/readdir") {
		url = url[len("/readdir"):]
		data, err := fsReadDir(url, h.dht)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			log.Println(err)
		} else {
			rw.WriteHeader(http.StatusOK)
			rw.Write(data)
		}
	} else if strings.HasPrefix(url, "/read") {
		url = url[len("/read"):]
		data, err := fsRead(url, 0, -1, h.dht)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			log.Println(err)
		} else {
			rw.WriteHeader(http.StatusOK)
			rw.Write(data)
		}
	} else if strings.HasPrefix(url, "/write") {
		url = url[len("/write"):]

		indata, err := ioutil.ReadAll(req.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			log.Println(err)
		}

		err = fsWrite(url, 0, indata, h.dht)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(err.Error()))
			log.Println(err)
		} else {
			rw.WriteHeader(http.StatusOK)
			rw.Write([]byte("Successfully uploaded"))
		}
	}
}

func (h *proxyHandler) uploadHandler(rw http.ResponseWriter, req *http.Request) {
	// Upload file.
	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte("Error while reading body of request: " + err.Error()))
		return
	}

	err = uploadFile(req.URL.Path, data, h.dht)
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		rw.Write([]byte(err.Error()))
	} else {
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte("File uploaded at " + req.URL.Path))
	}
}

func (h *proxyHandler) filesHandler(rw http.ResponseWriter, req *http.Request) {
	// Download file.
	if req.Method == "GET" {
		// Fetch file.
		data, err := downloadFile(req.URL.Path, h.dht)
		if err != nil {
			rw.WriteHeader(http.StatusNotFound)
			rw.Write([]byte(err.Error()))
			log.Println(err)
		} else {
			dst := make([]byte, base64.RawStdEncoding.DecodedLen(len(data)))
			base64.RawStdEncoding.Decode(dst, data)

			rw.WriteHeader(http.StatusOK)
			rw.Write(dst)
		}
	} else {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte("Unsupported method. Allowed methods: GET"))
		if err != nil {
			log.Println(err)
		}
	}
}

func (h *proxyHandler) apiHandler(rw http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		rw.WriteHeader(http.StatusBadRequest)
		_, err := rw.Write([]byte("Unsupported method. Allowed methods: GET"))
		if err != nil {
			log.Println(err)
		}
		return
	}
	v := h.dht.Respond("info", nil)
	rw.WriteHeader(http.StatusOK)
	rw.Write(v)
}
