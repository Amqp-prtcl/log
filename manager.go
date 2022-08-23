package log

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// ErrOutputClosed can be returned from Log call to outputs
// to signify that they should be removed from log manager
//
// an other error returns by log calls to outputs are ignored
//
// NOTE: upon ErrOutputClosed, log manager will remove output
// without calling LogClose
var ErrOutputClosed = fmt.Errorf("output closed")

var ErrManagerClosed = fmt.Errorf("log manager is closed")

type manager struct {
	outputs []Output
	mu      sync.Mutex

	b  atomic.Value
	ch chan managerEntry
}

type managerEntry struct {
	entry *LogEntry
	done  chan struct{}
}

func newManager() *manager {
	m := &manager{
		outputs: []Output{},
		mu:      sync.Mutex{},
		b:       atomic.Value{},
		ch:      make(chan managerEntry, 10),
	}
	m.b.Store(false)
	go func(m *manager) {
		var rm = make([]int, 0, 10)
		var r = 0
		for e := range m.ch {
			if e.entry != nil {
				rm = rm[:0]
				r = 0
				//st := time.Now()
				m.mu.Lock()
				for i, output := range m.outputs {
					if output.Log(e.entry) == ErrOutputClosed {
						rm = append(rm, i)
					}
				}
				for _, i := range rm {
					m.outputs[i-r] = m.outputs[len(m.outputs)-1]
					m.outputs = m.outputs[:len(m.outputs)-1]
					r++
				}
				m.mu.Unlock()

				for _, entry := range e.entry.Compiled {
					putBuf(entry.Buf)
				}
				//fmt.Println(time.Since(st))
			}
			if e.done != nil {
				close(e.done)
			}
		}
	}(m)

	return m
}

func (m *manager) addOutput(o Output) {
	if b, ok := m.b.Load().(bool); !ok || b {
		return
	}
	m.mu.Lock()
	m.outputs = append(m.outputs, o)
	o.OnAdd()
	m.mu.Unlock()
}

// remove output wont trigger LogClose.
// It is the caller's responsibility to close output
func (m *manager) removeOutput(o Output) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, v := range m.outputs {
		if v == o {
			m.outputs[i] = m.outputs[len(m.outputs)-1]
			m.outputs = m.outputs[:len(m.outputs)-1]
			return true
		}
	}
	return false
}

func (m *manager) forEach(fn func(int, Output)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i, v := range m.outputs {
		fn(i, v)
	}
}

// if entry is nil and block is true, log will wait for queue to finish
func (m *manager) log(entry *LogEntry, block bool) error {
	var e = managerEntry{
		entry: entry,
		done:  nil,
	}
	if block {
		e.done = make(chan struct{})
	}
	if b, ok := m.b.Load().(bool); !ok || b {
		return ErrManagerClosed
	}
	m.ch <- e
	if block {
		<-e.done
	}
	return nil
}

// Close waits for All Logs to finish before closing all Outputs
func (m *manager) Close() error {
	var entry = managerEntry{
		entry: nil,
		done:  make(chan struct{}),
	}
	if b, ok := m.b.Load().(bool); !ok || b {
		return ErrManagerClosed
	}
	m.b.Store(true)
	m.ch <- entry
	close(m.ch)
	<-entry.done

	m.mu.Lock()
	// TODO add multi errors
	for _, v := range m.outputs {
		v.LogClose()
	}
	m.outputs = []Output{}
	m.mu.Unlock()
	return nil
}

// see Logger.Lock() 's doc
func (m *manager) Lock() {
	m.mu.Lock()
}

func (m *manager) Unlock() {
	m.mu.Unlock()
}
