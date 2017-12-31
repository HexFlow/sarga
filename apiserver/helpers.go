package apiserver

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

	"github.com/sakshamsharma/sarga/common/dht"
)

const ChunkSizeBytes = 1024 * 1024 // 1 MB

func uploadFile(fileName string, data []byte, dht dht.DHT) error {
	type dataChunk struct {
		key             string
		data            []byte
		keyWithDataHash string
	}

	var chunks []dataChunk
	var err error
	dataLen := len(data)
	count := 0
	chunkCount := 0

	for count < dataLen {
		chunkCount++
		thisChunkLen := min(ChunkSizeBytes, dataLen-count)
		thisChunkHash := hashStr(fileName + "#" + strconv.Itoa(chunkCount))
		dataChunk := dataChunk{
			key:  thisChunkHash,
			data: data[count : count+thisChunkLen],
		}
		dataChunk.keyWithDataHash = hashStr(dataChunk.key + base64.StdEncoding.EncodeToString(dataChunk.data))
		chunks = append(chunks, dataChunk)
		count += thisChunkLen
	}

	if chunkCount == 1 {
		// Store the chunk directly, and prepend a 0 byte in the beginning to mark
		// that the complete file is in this data piece.
		err = dht.StoreValue(hashStr(fileName), append([]byte{0}, chunks[0].data...))
	} else {
		listOfChunkHashes := ""
		for i, chunk := range chunks {
			if i != 0 {
				listOfChunkHashes += "#"
			}
			listOfChunkHashes += chunk.keyWithDataHash
		}
		err = dht.StoreValue(hashStr(fileName), append([]byte{1}, []byte(listOfChunkHashes)...))

		for _, chunk := range chunks {
			if err != nil {
				return err
			}
			dataToStore := append([]byte{0}, []byte(chunk.data)...)
			err = dht.StoreValue(chunk.keyWithDataHash, dataToStore)
		}
	}
	return err
}

func downloadFile(fileName string, dht dht.DHT) ([]byte, error) {
	data, err := dht.FindValue(hashStr(fileName))
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty data returned by DHT for root key %q", hashStr(fileName))
	}

	if data[0] == 0 {
		// This is the whole data.
		return data[1:], nil
	}

	result := []byte{}
	for _, chunkHash := range bytes.Split(data[1:], []byte("#")) {
		chunk, err := dht.FindValue(string(chunkHash))
		if err != nil {
			return nil, err
		}
		if len(chunk) == 0 {
			return nil, fmt.Errorf("empty data returned by DHT for key %q", string(chunkHash))
		}
		result = append(result, chunk[1:]...)
	}

	return result, nil
}

func hashStr(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	return hex.EncodeToString(h.Sum(nil))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a < b {
		return b
	}
	return a
}
