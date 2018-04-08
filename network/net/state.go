package net

import (
	"../errors"
	"../utils"
	"github.com/boltdb/bolt"
	"fmt"
	"strconv"
	"crypto/sha256"
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
	fmt.Printf("read: bucket: %s\n", string(state.Bucket))
	keys, values := new(Keys), new(Values)

	state.Db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(state.Bucket).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			*keys = append(*keys, string(k))
			*values = append(*values, string(v))
			// fmt.Printf("read: key=%s, value=%s\n", k, v)
		}
		keysSha := sha256.Sum256([]byte(fmt.Sprintf("%v", keys)))
		valuesSha := sha256.Sum256([]byte(fmt.Sprintf("%v", keys)))
		fmt.Printf("read: keys=%s, values=%s\n", fmt.Sprintf("%x", keysSha[:6]), fmt.Sprintf("%x", valuesSha[:6]))

		return nil
	})

	return *keys, *values
}

func (state *State) Diff(state2 *State) (diff Diff) {
	diff = Diff{*state, *state2, map[utils.Any]utils.Any{}}
	keys, values := state.read()
	keys2, _ := state2.read()
	// fmt.Println("diff: keys:", len(keys))
	// fmt.Println("diff: keys2:", len(keys2))

	unique := sliceDiff(keys, keys2)
	diff.populate(unique, values)

	return diff
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

	// fmt.Println("diff:", diff)
	return diff
}
