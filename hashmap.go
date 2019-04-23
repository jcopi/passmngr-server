package main

import "bitbucket.org/creachadair/cityhash"

type HashmapNode struct {
	Key   []byte
	Value interface{}
	dist  uint8
	used  bool
}

type Hashmap struct {
	table    []HashmapNode
	length   uint32
	maxLoad  float32
	maxProbe uint8
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

	hm.table = make([]HashmapNode, 0, hmInitSize)
	hm.length = 0
	hm.maxProbe = fastLog2(uint64(cap(hm.table)))

	return hm
}

func (hm *Hashmap) grow() {
	newCap := nextHashMapSize(cap(hm.table))
	tmp := make([]HashmapNode, 0, newCap)
	hm.rehash(tmp)
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

func (hm *Hashmap) probe(hash uint32, key []byte, value interface{}, bool altSize) {
	idx := fastRange32(hash, uint32(cap(hm.table)))
	i, j, spi, probeCount := idx, 0, idx, uint8(0)
	smallerProbe, found := false, false
	for probeCount < hm.maxProbe && !found {
		j = i
	}
}

func (hm *Hashmap) Set(key []byte, val interface{}) {
	if float32(cap(hm.table))*hm.maxLoad > float32(hm.length+1) {
		hm.grow()
		hm.Set(key, val)
	} else {
		hash := cityhash.Hash32(key)
		hm.probe(hash, key, value, true)
	}
}
