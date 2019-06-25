package main

import (
	"bytes"

	"github.com/creachadair/cityhash"
)

type Hashable interface {
	Hash() uint32
	Equal(Hashable) bool
	AsBytes() []byte
}

type HashableByteSlice []byte
type HashableUint64 uint64

func (b HashableByteSlice) Hash() uint32 {
	return cityhash.Hash32(b)
}
func (b HashableByteSlice) Equal(c Hashable) bool {
	if b == nil && c == nil {
		return true
	} else if b == nil || c == nil {
		return false
	}
	return bytes.Equal(b.AsBytes(), c.AsBytes())
}

func (b HashableByteSlice) AsBytes() []byte {
	return []byte(b)
}

func (u HashableUint64) Hash() uint32 {
	return uint32(u ^ 9223372036854775783)
}

func (u HashableUint64) Equal(c Hashable) bool {
	return bytes.Equal(u.AsBytes(), c.AsBytes())
}

func (u HashableUint64) AsBytes() []byte {
	return []byte{
		byte(u >> 56),
		byte((u & 0x00FF000000000000) >> 48),
		byte((u & 0x0000FF0000000000) >> 40),
		byte((u & 0x000000FF00000000) >> 32),
		byte((u & 0x00000000FF000000) >> 24),
		byte((u & 0x0000000000FF0000) >> 16),
		byte((u & 0x000000000000FF00) >> 8),
		byte((u & 0x00000000000000FF)),
	}
}

type HashmapNode struct {
	Key   Hashable
	Value interface{}
	dist  uint8
	used  bool
}

type Hashmap struct {
	table       []HashmapNode
	length      uint32
	maxLoad     float32
	maxProbe    uint8
	MaxCapacity int
}

const (
	hmInitSize       int     = 16
	hmDefaultMaxLoad float32 = 0.75
)

func fastLog2(n uint64) uint8 {
	var i uint8
	if n == 0 {
		i = 1
	}
	if n >= (uint64(1) << 32) {
		i += uint8(32)
		n >>= 32
	}
	if n >= (uint64(1) << 16) {
		i += uint8(16)
		n >>= 16
	}
	if n >= (uint64(1) << 8) {
		i += uint8(8)
		n >>= 8
	}
	if n >= (uint64(1) << 4) {
		i += uint8(4)
		n >>= 4
	}
	if n >= (uint64(1) << 2) {
		i += uint8(2)
		n >>= 2
	}
	if n >= (uint64(1) << 1) {
		i += uint8(1)
		n >>= 1
	}
	return i
}

func nextHashMapSize(c int) (i int) {
	i = c + (c >> 1)
	return i
}

func fastRange32(x uint32, n uint32) uint32 {
	return uint32(uint64(x) * uint64(n) >> 32)
}

func NewHashmap() Hashmap {
	var hm Hashmap

	hm.table = make([]HashmapNode, hmInitSize, hmInitSize)
	hm.length = 0
	hm.maxProbe = fastLog2(uint64(cap(hm.table)))
	hm.MaxCapacity = -1

	return hm
}

func (hm *Hashmap) grow() bool {
	if hm.MaxCapacity >= 0 && cap(hm.table) >= hm.MaxCapacity {
		// Can't grow, at max size
		return false
	}
	newCap := nextHashMapSize(cap(hm.table))
	tmp := make([]HashmapNode, newCap, newCap)
	hm.rehash(tmp)

	return true
}

func (hm *Hashmap) rehash(newSlice []HashmapNode) {
	oldSlice := hm.table
	hm.table = newSlice
	hm.length = 0
	hm.maxProbe = fastLog2(uint64(cap(hm.table)))

	for _, n := range oldSlice {
		if n.used {
			hm.Set(n.Key, n.Value)
		}
	}
}

func (hm *Hashmap) probe(hash uint32, key Hashable, value interface{}, altSize bool) bool {
	idx := fastRange32(hash, uint32(cap(hm.table)))
	i, j, spi, probeCount := idx, uint32(0), idx, uint8(0)
	smallerProbe, found := false, false
	for probeCount < hm.maxProbe && !found {
		if j = i; i >= uint32(cap(hm.table)) {
			j -= uint32(cap(hm.table))
		}

		if !smallerProbe && (hm.table[j].dist < probeCount || !hm.table[j].used) {
			smallerProbe = true
			spi = j
		}

		if key.Equal(hm.table[j].Key) {
			found = true
			spi = j
			break
		}

		probeCount++
		i++
	}
	if found {
		hm.table[spi].Value = value
	} else if !smallerProbe {
		if !hm.grow() {
			return false
		}
		hm.probe(hash, key, value, true)
	} else {
		if hm.table[spi].used {
			newHash := hm.table[spi].Key.Hash()
			newKey := hm.table[spi].Key
			newValue := hm.table[spi].Value
			hm.table[spi].Key = key
			hm.table[spi].Value = value
			hm.table[spi].dist = uint8(spi - idx)
			if altSize {
				hm.length++
			}
			hm.probe(newHash, newKey, newValue, false)
		} else {
			hm.table[spi].Key = key
			hm.table[spi].Value = value
			hm.table[spi].dist = uint8(spi - idx)
			hm.table[spi].used = true
			if altSize {
				hm.length++
			}
		}
	}

	return true
}

func (hm *Hashmap) Set(key Hashable, value interface{}) bool {
	if float32(cap(hm.table))*hm.maxLoad > float32(hm.length+1) {
		if !hm.grow() {
			return false
		}
		hm.Set(key, value)
	} else {
		hash := key.Hash()
		if !hm.probe(hash, key, value, true) {
			return false
		}
	}

	return true
}

func (hm *Hashmap) Delete(key Hashable) (ok bool, value interface{}) {
	hash := key.Hash()
	i, j := fastRange32(hash, uint32(cap(hm.table))), uint32(0)
	probeCount := uint8(0)

	for probeCount < hm.maxProbe {
		if j = i; j >= uint32(cap(hm.table)) {
			j -= uint32(cap(hm.table))
		}

		if key.Equal(hm.table[j].Key) {
			value = hm.table[j].Value
			var tmp HashmapNode
			hm.table[j] = tmp
			hm.length--
			ok = true
			break
		}

		i++
		probeCount++
	}

	return
}

func (hm *Hashmap) Get(key Hashable) interface{} {
	var value interface{}
	hash := key.Hash()
	i, j := fastRange32(hash, uint32(cap(hm.table))), uint32(0)
	probeCount := uint8(0)

	for probeCount < hm.maxProbe {
		if j = i; j >= uint32(cap(hm.table)) {
			j -= uint32(cap(hm.table))
		}

		if key.Equal(hm.table[j].Key) {
			value = hm.table[j].Value
			break
		}

		i++
		probeCount++
	}

	return value
}

func (hm *Hashmap) GetPreHash(key Hashable, hash uint32) interface{} {
	var value interface{}
	i, j := fastRange32(hash, uint32(cap(hm.table))), uint32(0)
	probeCount := uint8(0)

	for probeCount < hm.maxProbe {
		if j = i; j >= uint32(cap(hm.table)) {
			j -= uint32(cap(hm.table))
		}

		if key.Equal(hm.table[j].Key) {
			value = hm.table[j].Value
			break
		}

		i++
		probeCount++
	}

	return value
}
