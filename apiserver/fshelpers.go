package apiserver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sakshamsharma/sarga/common/dht"
)

var Splitter = []byte{':', ':'}

type FileType int

const (
	FileChunkListT FileType = iota // List of hashes of chunks of this file
	FileChunkT                     // An actual chunk of a file
	DirectoryT                     // A directory with a list of sub-files
)

type dhtEntryFormat struct {
	Key        string // Hash using which this file can be found
	Parent     string // Useful for file chunks to purge old versions
	Metadata   string // Could contain application specific metadata
	FileType   FileType
	Data       []byte
	LastModify int
	LastRead   int
}

func curUNIXTime() int {
	// Fix this. Make this UTC.
	return time.Now().Second()
}

func getItem(path string, dht dht.DHT) (*dhtEntryFormat, error) {
	return getItemFromHash(hashStr(path), dht)
}

func getItemFromHash(hash string, dht dht.DHT) (*dhtEntryFormat, error) {
	data, err := dht.FindValue(hash)
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("empty data returned by DHT for root key %q", hash)
	}

	entry := &dhtEntryFormat{}
	err = json.Unmarshal(data, &entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func addItemToDirectory(dirname, filename string, dht dht.DHT) error {
	entry, err := getItem(dirname, dht)
	if err != nil {
		fileTime := curUNIXTime()
		entry = &dhtEntryFormat{
			Key:        hashStr(dirname),
			Parent:     dirname, // Parent is the raw path in case of FileChunkListT
			Metadata:   "",
			FileType:   DirectoryT,
			Data:       []byte{},
			LastModify: fileTime,
			LastRead:   fileTime,
		}
	}

	if entry.FileType != DirectoryT {
		return fmt.Errorf("%q is not a directory", dirname)
	}

	files := []string{}
	for _, file := range bytes.Split(entry.Data, Splitter) {
		if len(file) != 0 {
			files = append(files, string(file))
		}
	}
	files = append(files, filename)

	entry.Data = []byte(strings.Join(files, string(Splitter)))

	updatedEntry, _ := json.Marshal(entry)

	return dht.StoreValue(hashStr(dirname), updatedEntry)
}

func fsGetAttr(path string, dht dht.DHT) ([]byte, error) {
	if path == "/" {
		resp, _ := json.Marshal(getAttrResp{
			FileType: DirectoryT,
		})
		return resp, nil
	}

	entry, err := getItem(path, dht)
	if err != nil {
		return nil, err
	}

	resp, _ := json.Marshal(getAttrResp{
		FileType: entry.FileType,
	})
	return resp, nil
}

func fsReadDir(path string, dht dht.DHT) ([]byte, error) {
	entry, err := getItem(path, dht)
	if err != nil {
		if path == "/" {
			resp, _ := json.Marshal(readDirResp{
				Files: []string{},
			})
			return resp, nil
		}

		return nil, err
	}

	if entry.FileType != DirectoryT {
		return nil, fmt.Errorf("%q is not a directory", path)
	}

	files := []string{}
	for _, file := range bytes.Split(entry.Data, Splitter) {
		files = append(files, string(file))
	}

	resp, _ := json.Marshal(readDirResp{
		Files: files,
	})
	return resp, nil
}

func fsRead(path string, offset int, length int, dht dht.DHT) ([]byte, error) {
	entry, err := getItem(path, dht)
	if err != nil {
		return nil, err
	}

	// TODO: Confirm if this check is fine
	if entry.FileType != FileChunkListT {
		return nil, fmt.Errorf("%q is not a readable file", path)
	}
	fmt.Println("Got file chunk list", hashStr(path))

	chunkHashes := bytes.Split(entry.Data, []byte(Splitter))

	startChunk := offset / ChunkSizeBytes
	remaining := length

	if remaining == -1 {
		// Max value of integer
		remaining = int((^uint(0)) >> 1)
	}

	resp := readResp{}
	for chunkID := startChunk; chunkID <= len(chunkHashes)-1 && remaining > 0; chunkID++ {
		chunk, err := getItemFromHash(string(chunkHashes[chunkID]), dht)
		if err != nil {
			return nil, err
		}

		if chunk.FileType != FileChunkT {
			return nil, fmt.Errorf("%q is not a file chunk", chunkHashes[chunkID])
		}

		startByte := max(0, offset-chunkID*ChunkSizeBytes)
		endByte := min(len(chunk.Data), startByte+remaining)
		remaining -= (endByte - startByte)

		dd := chunk.Data[startByte:endByte]
		resp = append(resp, dd...)
	}

	// For information on why we did not do json.Marshal here:
	// https://stackoverflow.com/questions/36465065/
	// https://stackoverflow.com/questions/24229205/
	return resp, nil
}

func fsWrite(path string, offset int, data []byte, dht dht.DHT) error {
	if offset != 0 {
		return fmt.Errorf("non zero offsets not supported for fsWrite")
	}

	type dataChunk struct {
		key  string
		data []byte
	}

	var chunks []dataChunk
	dataLen := len(data)
	count := 0
	chunkCount := 0

	for count < dataLen {
		chunkCount++
		thisChunkLen := min(ChunkSizeBytes, dataLen-count)
		thisChunkHash := hashStr(path + "#" + strconv.Itoa(chunkCount))
		chunks = append(chunks, dataChunk{
			key:  thisChunkHash,
			data: data[count : count+thisChunkLen],
		})
		count += thisChunkLen
	}

	listOfChunkHashes := ""
	for i, chunk := range chunks {
		if i != 0 {
			listOfChunkHashes += "#"
		}
		listOfChunkHashes += chunk.key
	}

	fileTime := curUNIXTime()
	entry, _ := json.Marshal(dhtEntryFormat{
		Key:        hashStr(path),
		Parent:     path, // Parent is the raw path in case of FileChunkListT
		Metadata:   "",
		FileType:   FileChunkListT,
		Data:       []byte(listOfChunkHashes),
		LastModify: fileTime,
		LastRead:   fileTime,
	})

	err := dht.StoreValue(hashStr(path), entry)
	fmt.Println("Storing file chunk list", hashStr(path))

	for _, chunk := range chunks {
		if err != nil {
			return err
		}
		dataToStore, _ := json.Marshal(dhtEntryFormat{
			Key:        chunk.key,
			Parent:     path,
			Metadata:   "",
			FileType:   FileChunkT,
			Data:       chunk.data,
			LastModify: fileTime,
			LastRead:   fileTime,
		})
		fmt.Println("Storing chunk", chunk.key)
		err = dht.StoreValue(chunk.key, dataToStore)
	}

	if err != nil {
		return err
	}

	// Now add this file to its parent folder
	pathChunks := strings.Split(path, "/")
	pathChunkLen := len(pathChunks)
	dirpath := strings.Join(pathChunks[:pathChunkLen-1], "/")
	filename := pathChunks[pathChunkLen-1]

	return addItemToDirectory("/"+dirpath, filename, dht)
}
