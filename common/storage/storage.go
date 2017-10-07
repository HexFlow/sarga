package storage

type Storage interface {
	Init()
	Store(key string, data []byte) error
	Fetch(key string) ([]byte, error)
	Empty() error
}