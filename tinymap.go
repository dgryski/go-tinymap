// tinymap is a compact map for keys less than 64
package tinymap

import (
	"math/bits"
)

// Map is a compact map
type Map struct {
	bits uint64
	data []uint64
}

// Insert adds a key to the map.
func (m *Map) Insert(k uint8, val uint64) {
	mask := uint64(1) << k
	idx := bits.OnesCount64(m.bits & (mask - 1))
	if m.bits&mask != 0 {
		m.data[idx] = val
		return
	}

	m.bits |= mask
	m.data = append(m.data, 0)
	copy(m.data[idx+1:], m.data[idx:])
	m.data[idx] = val
}

// Delete removes a key from the map.
func (m *Map) Delete(k uint8) {
	mask := uint64(1) << k
	if m.bits&mask == 0 {
		return
	}

	m.bits &^= mask
	idx := bits.OnesCount64(m.bits & (mask - 1))
	copy(m.data[idx:], m.data[idx+1:])
	m.data = m.data[:len(m.data)-1]
}

// Lookup returns the value from a map and boolean if it was found.
func (m *Map) Lookup(k uint8) (val uint64, ok bool) {
	mask := uint64(1) << k
	if m.bits&mask == 0 {
		return 0, false
	}

	idx := bits.OnesCount64(m.bits & (mask - 1))
	return m.data[idx], true
}
