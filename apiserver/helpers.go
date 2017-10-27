package apiserver

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"strconv"

	"github.com/sakshamsharma/sarga/common/dht"
)

const ChunkSizeBytes = 1024 * 1024

type dataChunk struct {
	key  string
	data []byte
}

func hashStr(s string) string {
	return hex.EncodeToString(sha1.New().Sum([]byte(s)))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func uploadFile(fileName string, data []byte, dht dht.DHT) error {

	var chunks []dataChunk
	var err error
	dataLen := len(data)
	count := 0
	chunkCount := 1

	for count < dataLen {
		thisChunkLen := min(ChunkSizeBytes, dataLen-count)
		chunks = append(chunks, dataChunk{
			key:  hashStr(fileName + "#" + strconv.Itoa(chunkCount)),
			data: data[count : count+thisChunkLen],
		})
		count += thisChunkLen
		chunkCount++
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
			listOfChunkHashes += chunk.key
		}
		err = dht.StoreValue(hashStr(fileName), append([]byte{1}, []byte(listOfChunkHashes)...))

		for _, chunk := range chunks {
			if err != nil {
				return err
			}
			err = dht.StoreValue(hashStr(chunk.key), append([]byte{0}, []byte(chunk.data)...))
		}
	}
	return err
}

func downloadFile(fileName string, dht dht.DHT) ([]byte, error) {
	data, err := dht.FindValue(hashStr(fileName))
	if err != nil {
		return nil, err
	}

	if data[0] == 0 {
		// This is the whole data.
		return data[1:], nil
	}

	result := []byte{}
	for _, chunkHash := range bytes.Split(data[1:], []byte("#")) {
		chunk, err := dht.FindValue(string(chunkHash))
		if err == nil {
			return nil, err
		}
		result = append(result, chunk[1:]...)
	}

	return result, nil
}
