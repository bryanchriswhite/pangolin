package net

import (
	"../errors"
	"../utils"
	"fmt"
	"github.com/boltdb/bolt"
)

// type TxId int
//
// type Tx struct {
// 	Id TxId
// }
type Keys []utils.Any
type Values []utils.Any

type State struct {
	Db     *bolt.DB
	Bucket []byte
}

type Diff struct {
	state1 State
	state2 State
	data   map[utils.Any]utils.Any
	// data   []Tx
}

// TODO: return *[]TxId, *[]Tx
func (state *State) read(state2 *State) (Keys, Values) {
	keys, values := new(Keys), new(Values)

	state.Db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(state.Bucket).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			*keys = append(*keys, string(k))
			*values = append(*values, string(v))
			// fmt.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})

	return *keys, *values
}

func (state *State) diff(state2 *State) (diff, diff2 Diff) {
	diff = Diff{*state, *state2, map[utils.Any]utils.Any{}}
	diff2 = Diff{*state2, *state, map[utils.Any]utils.Any{}}
	keys, values := state.read(state)
	keys2, values2 := state.read(state2)

	unique1, unique2 := sliceDiffs(keys, keys2)
	diff.populate(unique1, values)
	diff2.populate(unique2, values2)

	for _, elem := range unique2 {
		diff2.data[elem.Value] = values[elem.Index]
	}

	fmt.Println(keys, values)

	return diff, diff2
}

func (state *State) write(diff *Diff) {
	if len(diff.data) == 0 {
		return
	}

	state.Db.Update(func(tx *bolt.Tx) (err error) {
		err = error(nil)
		b := tx.Bucket(state.Bucket)

		for key, value := range diff.data {
			err = coerce(key, value, func(k, v []byte) {
				b.Put(k, v)
			})

		}

		return err
	})
	fmt.Printf("updating State: %v\nwith diff: %v", state, diff)
}

func coerce(key, value utils.Any, f func(k, v []byte)) (err error) {
	var _key, _value []byte

	switch k := key.(type) {
	case string:
		_key = []byte(k)
	default:
		err = errors.NewCoercionError(k)
	}

	switch v := value.(type) {
	case string:
		_value = []byte(v)
	default:
		err = bolt.ErrIncompatibleValue
	}

	if err == nil {
		f(_key, _value)
	}

	return err
}

type elem struct {
	Index  int
	Value  utils.Any
	unique bool
}

func (d *Diff) populate(unique []elem, values Values) {
	for _, elem := range unique {
		d.data[elem.Value] = values[elem.Index]
	}
}

func sliceDiffs(slice, slice2 []utils.Any) (diff1, diff2 []elem) {
	diff1 = sliceDiff(slice, slice2)
	diff2 = sliceDiff(slice2, slice)

	return diff1, diff2
}

func sliceDiff(slice, slice2 []utils.Any) (diff []elem) {
	m := map[utils.Any]elem{}

	for i, v := range slice {
		m[v] = elem{i, v, true}
	}
	for _, v := range slice2 {
		e, ok := m[v]
		if ok {
			e.unique = false
			break
		}
	}

	for _, e := range m {
		if e.unique == true {
			diff = append(diff, e)
		}
	}

	return diff
}
