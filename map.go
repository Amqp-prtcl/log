package log

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// M is an JSON marshalable Ordered Map
// meaning that first field is marshaled first, second is marshaled second, ect...
//
// Note that M can have duplicate Keys
//
// M is not safe by itself for concurrent use
type M []MapEntry

type MapEntry struct {
	Key string
	Val interface{}
}

func (e MapEntry) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(e.Val)
	if err != nil {
		return nil, err
	}
	return []byte(fmt.Sprintf(`{"%s": %s}`, e.Key, data)), nil
}

func (m *M) Add(key string, val interface{}) {
	*m = append(*m, MapEntry{key, val})
}

// There can be duplicate keys
func (m *M) AddE(e ...MapEntry) {
	*m = append(*m, e...)
}

// only removes first occurrence
func (m *M) Remove(key string) (interface{}, bool) {
	for i, entry := range *m {
		if entry.Key == key {
			(*m)[i] = (*m)[len(*m)-1]
			*m = (*m)[:len(*m)-1]
			return entry, true
		}
	}
	return nil, false
}

func (m *M) ForEach(fn func(int, MapEntry) bool) {
	for i, v := range *m {
		if fn(i, v) {
			break
		}
	}
}

// return M so that it is marshaled as a JSON array of JSON object
// with each field being an object
func (m M) AsArray() []MapEntry {
	return m
}

func (m M) MarshalJSON() ([]byte, error) {
	var buf = []byte{'{'}
	for i, entry := range m {
		data, err := json.Marshal(entry.Val)
		if err != nil {
			return buf, err
		}
		buf = append(buf, strconv.Quote(entry.Key)+":"...)
		buf = append(buf, data...)
		if i < len(m)-1 {
			buf = append(buf, ',')
		}
	}
	buf = append(buf, '}')
	return buf, nil
}
