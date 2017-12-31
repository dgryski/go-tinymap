package tinymap

import (
	"encoding/binary"
	"testing"
	"testing/quick"

	ddmin "github.com/dgryski/go-ddmin"
)

type opType uint8

const (
	opInsert opType = 0
	opDelete opType = 1
	opLookup opType = 2
)

type action struct {
	op opType
	k  uint8
	v  uint64
}

func getuint(d []byte) (uint64, []byte) {
	switch len(d) {
	case 0:
		return 0, nil
	case 1:
		return uint64(d[0]), d[1:]
	case 2, 3:
		return uint64(binary.LittleEndian.Uint16(d)), d[2:]
	default:
		return uint64(binary.LittleEndian.Uint32(d)), d[4:]
	}
}

func makeOps(d []byte) []action {
	var actions []action

	for len(d) > 0 {
		op, k := opType(d[0]>>6), d[0]&0x3f
		d = d[1:]

		switch op {
		case opDelete, opLookup:
			actions = append(actions, action{op: op, k: k})
		case opInsert:
			var v uint64
			v, d = getuint(d)
			actions = append(actions, action{op: op, k: k, v: v})
		default:
			// nothing
		}
	}

	return actions
}

func TestMap(t *testing.T) {

	var verbose bool

	f := func(data []byte) (ok bool) {

		defer func() {
			if r := recover(); r != nil {
				if verbose {
					t.Logf("recovered: %#v", r)
				}
				ok = false
			}
		}()

		actions := makeOps(data)

		var m Map
		r := make(map[uint8]uint64)

		for _, action := range actions {

			switch action.op {
			case opDelete:
				if verbose {
					t.Logf("delete %v", action.k)
				}
				m.Delete(action.k)
				delete(r, action.k)
			case opInsert:
				if verbose {
					t.Logf("insert %v %v", action.k, action.v)
				}

				m.Insert(action.k, action.v)
				r[action.k] = action.v
			case opLookup:
				if verbose {
					t.Logf("lookup %v", action.k)
				}
				mval, mok := m.Lookup(action.k)
				rval, rok := r[action.k]
				if mval != rval || mok != rok {
					if verbose {
						t.Logf("lookup check: k=%v mval=%v rval=%v mok=%v rok=%v", action.k, mval, rval, mok, rok)
					}
					return false
				}
			}
		}

		// check remaining state
		for i := uint8(0); i < 64; i++ {
			mval, mok := m.Lookup(i)
			rval, rok := r[i]
			if mval != rval || mok != rok {
				if verbose {
					t.Logf("final check: k=%v mval=%v rval=%v mok=%v rok=%v", i, mval, rval, mok, rok)
				}
				return false
			}
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		data := err.(*quick.CheckError).In[0].([]byte)
		t.Error(err)
		data = ddmin.Minimize(data, func(d []byte) ddmin.Result {
			if !f(d) {
				return ddmin.Fail
			}

			return ddmin.Pass
		})
		t.Logf("reduced: %v", data)
		verbose = true
		f(data)
	}
}
