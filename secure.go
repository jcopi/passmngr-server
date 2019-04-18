package main

import (
	"crypto/rand"
	"encoding/binary"
	"sync"
	"time"
)

type EphemeralRatchet struct {
	key     []byte
	context []byte
	time    time.Time
}

func NewEphemeralRatchet(key []byte, context []byte) EphemeralRatchet {
	return EphemeralRatchet{key, context, time.Now()}
}

type EphemeralRatchetStorage struct {
	storage map[uint64]EphemeralRatchet
	lock    *sync.Mutex
}

func NewEphemeralRatchetStorage() (e EphemeralRatchetStorage) {
	e.lock = &sync.Mutex{}
	return
}

func (s *EphemeralRatchetStorage) Enqueue(e EphemeralRatchet) (u uint64) {
	s.lock.Lock()
	defer s.lock.Unlock()

	tmp := [8]byte{}
	rand.Read(tmp[:])
	u = binary.LittleEndian.Uint64(tmp[:])
	for _, ok := s.storage[u]; ok; _, ok = s.storage[u] {
		rand.Read(tmp[:])
		u = binary.LittleEndian.Uint64(tmp[:])
	}

	s.storage[u] = e

	return
}

func (s *EphemeralRatchetStorage) Dequeue(u uint64) (e EphemeralRatchet, ok bool) {
	s.lock.Lock()
	defer s.lock.Unlock()

	if e, ok = s.storage[u]; ok {
		delete(s.storage, u)
	}

	return
}
