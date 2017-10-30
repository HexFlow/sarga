package sdht

import "errors"

type Storage map[string][]byte

func (s Storage) Get(key string) ([]byte, error) {
	if val, ok := s[key]; ok {
		return val, nil
	}
	return nil, errors.New("value not found")
}

func (s Storage) Set(key string, val []byte) error {
	s[key] = val
	return nil
}

func (s Storage) Del(key string) error {
	delete(s, key)
	return nil
}

func (s Storage) Marshal() string {
	tmpMap := map[string]string{}
	for key, val := range s {
		tmpMap[key] = string(val)
	}
	return string(marshal(tmpMap))
}
