package net

import (
	"../errors"
	"../utils"
	"github.com/boltdb/bolt"
	"fmt"
	"strconv"
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
	Data   map[utils.Any]utils.Any
	// Data   []Tx
}

// TODO: return *[]TxId, *[]Tx
func (state *State) read() (Keys, Values) {
	fmt.Printf("read bucket: %s\n", string(state.Bucket))
	keys, values := new(Keys), new(Values)

	state.Db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(state.Bucket).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			*keys = append(*keys, string(k))
			*values = append(*values, string(v))
			// fmt.Printf("read: key=%s, value=%s\n", k, v)
		}
		fmt.Printf("read: keys=%v, values=%v\n", keys, values)

		return nil
	})

	return *keys, *values
}

func (state *State) Diff(state2 *State) (diff, diff2 Diff) {
	diff = Diff{*state, *state2, map[utils.Any]utils.Any{}}
	diff2 = Diff{*state2, *state, map[utils.Any]utils.Any{}}
	keys, values := state.read()
	keys2, values2 := state2.read()
	fmt.Println("keys:", len(keys))
	fmt.Println("keys2:", len(keys2))

	unique1, unique2 := sliceDiffs(keys, keys2)
	diff.populate(unique1, values)
	diff2.populate(unique2, values2)

	return diff, diff2
}

func (state *State) write(diff *Diff) (err error) {
	// fmt.Printf("updating bucket: %v\nwith Diff: %v\n\n", string(state.Bucket), Diff)
	state.Db.Update(func(tx *bolt.Tx) (err error) {
		b := tx.Bucket(state.Bucket)
		// fmt.Printf("write bucket: %s\n", string(state.Bucket))

		for key, value := range diff.Data {
			// fmt.Printf("key: %v\nvalue: %v\n", key, value)
			var boltErr error
			err = coerce(func(coercion [][]byte) {
				k, v := coercion[0], coercion[1]

				// fmt.Printf("write k: %v; v: %v\n", string(k), string(v))
				boltErr = b.Put(k, v)
			}, key, value)

			if boltErr != nil {
				return boltErr
			}

			if err != nil {
				return err
			}
		}

		// Database transaction is aborted if error is returned
		return err
	})

	return err
}

func coerce(f func([][]byte), values ...utils.Any) (err error) {
	_values := make([][]byte, 0)
	for _, value := range values {
		switch v := value.(type) {
		case string:
			_values = append(_values, []byte(v))
		case int64:
			_values = append(_values, []byte(strconv.FormatInt(v, 10)))
		default:
			err = errors.NewCoercionError(v)
			break
		}
	}

	if err == nil {
		f(_values)
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
		d.Data[elem.Value] = values[elem.Index]
	}
}

func (d *Diff) isEmpty() bool {
	return len(d.Data) == 0
}

func sliceDiffs(slice, slice2 []utils.Any) (diff1, diff2 []elem) {
	diff1 = sliceDiff(slice, slice2)
	diff2 = sliceDiff(slice2, slice)

	return diff1, diff2
}

func sliceDiff(slice, slice2 []utils.Any) (diff []elem) {
	// fmt.Println("slice:", slice)
	// fmt.Println("slice2:", slice2)
	m := map[utils.Any]*elem{}

	for i, v := range slice {
		m[v] = &elem{i, v, true}
	}

	for _, v := range slice2 {
		e, ok := m[v]
		if ok {
			e.unique = false
		}
	}

	for _, e := range m {
		if e.unique == true {
			diff = append(diff, *e)
		}
	}

	fmt.Println("diff:", diff)
	return diff
}
