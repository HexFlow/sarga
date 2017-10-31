package apiserver

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/sakshamsharma/sarga/common/dht"
	"github.com/sakshamsharma/sarga/common/iface"
	"github.com/sakshamsharma/sarga/impl/httpnet"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func TestUploadDownloadE2E(t *testing.T) {

	dht := &dht.FakeDHT{}
	dht.Init(iface.Address{}, []iface.Address{}, &httpnet.HTTPNet{})
	port := rand.Intn(5000) + 2000

	// TODO(sakshams): This goroutine does not terminate cleanly yet.
	go StartAPIServer(iface.CommonArgs{
		Port: port,
		IP:   "127.0.0.1",
	}, dht)

	time.Sleep(2)

	addr := "http://127.0.0.1:" + strconv.Itoa(port)

	testLen := rand.Intn(1024*20) + 3
	buf := make([]byte, testLen)
	rand.Read(buf)

	ss := base64.RawStdEncoding.EncodeToString(buf)
	bufReader := ioutil.NopCloser(bytes.NewBuffer([]byte(ss)))

	_, err := http.Post(addr+"/sarga/upload/coolfile", "text/plain", bufReader)
	if err != nil {
		t.Fatalf(err.Error())
	}

	resp, err := http.Get(addr + "/sarga/files/coolfile")
	if err != nil {
		t.Fatalf(err.Error())
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf(err.Error())
	}

	err = compareBufs(data, buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUploadDownload(t *testing.T) {
	dht := &dht.FakeDHT{}
	dht.Init(iface.Address{}, []iface.Address{}, &httpnet.HTTPNet{})

	testLen := rand.Intn(1024 * 20)
	buf := make([]byte, testLen)
	rand.Read(buf)

	err := uploadFile("coolfile", buf, dht)
	if err != nil {
		t.Fatal(err)
	}

	data, err := downloadFile("coolfile", dht)
	if err != nil {
		t.Fatal(err)
	}

	err = compareBufs(data, buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestChunking(t *testing.T) {
	dht := &dht.FakeDHT{}
	dht.Init(iface.Address{}, []iface.Address{}, &httpnet.HTTPNet{})

	testLen := 5*ChunkSizeBytes + rand.Intn(1024)

	buf := make([]byte, testLen)
	rand.Read(buf)

	err := uploadFile("coolfile", buf, dht)
	if err != nil {
		t.Fatal(err)
	}

	data, err := downloadFile("coolfile", dht)
	if err != nil {
		t.Fatal(err)
	}

	err = compareBufs(data, buf)
	if err != nil {
		t.Fatal(err)
	}
}

// compareBufs compares the expected buffer buf with the received buffer data.
func compareBufs(data, buf []byte) error {
	if len(data) != len(buf) {
		return fmt.Errorf("data of invalid length received, expected %d, got %d", len(buf), len(data))
	}

	for i := 0; i < len(buf); i++ {
		if data[i] != buf[i] {
			return fmt.Errorf("invalid data, byte at %d is different, expected %q, got %q", i, buf[i], data[i])
		}
	}
	return nil
}
