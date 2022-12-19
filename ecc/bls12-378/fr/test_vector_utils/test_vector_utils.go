// Copyright 2020 ConsenSys Software Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by consensys/gnark-crypto DO NOT EDIT

package test_vector_utils

import (
	"encoding/json"
	"fmt"
	"github.com/consensys/gnark-crypto/ecc/bls12-378/fr"
	"github.com/consensys/gnark-crypto/ecc/bls12-378/fr/polynomial"
	"hash"

	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type ElementTriplet struct {
	key1        fr.Element
	key2        fr.Element
	key2Present bool
	value       fr.Element
	used        bool
}

func (t *ElementTriplet) CmpKey(o *ElementTriplet) int {
	if cmp1 := t.key1.Cmp(&o.key1); cmp1 != 0 {
		return cmp1
	}

	if t.key2Present {
		if o.key2Present {
			return t.key2.Cmp(&o.key2)
		}
		return 1
	} else {
		if o.key2Present {
			return -1
		}
		return 0
	}
}

var MapCache = make(map[string]*ElementMap)

func ElementMapFromFile(path string) (*ElementMap, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if h, ok := MapCache[path]; ok {
		return h, nil
	}
	var bytes []byte
	if bytes, err = os.ReadFile(path); err == nil {
		var asMap map[string]interface{}
		if err = json.Unmarshal(bytes, &asMap); err != nil {
			return nil, err
		}

		var h ElementMap
		if h, err = CreateElementMap(asMap); err == nil {
			MapCache[path] = &h
		}

		return &h, err

	} else {
		return nil, err
	}
}

func CreateElementMap(rawMap map[string]interface{}) (ElementMap, error) {
	res := make(ElementMap, 0, len(rawMap))

	for k, v := range rawMap {
		var entry ElementTriplet
		if _, err := SetElement(&entry.value, v); err != nil {
			return nil, err
		}

		key := strings.Split(k, ",")
		switch len(key) {
		case 1:
			entry.key2Present = false
		case 2:
			entry.key2Present = true
			if _, err := SetElement(&entry.key2, key[1]); err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("cannot parse %T as one or two field elements", v)
		}
		if _, err := SetElement(&entry.key1, key[0]); err != nil {
			return nil, err
		}

		res = append(res, &entry)
	}

	res.sort()
	return res, nil
}

type ElementMap []*ElementTriplet

type MapHash struct {
	Map        *ElementMap
	state      fr.Element
	stateValid bool
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (m *MapHash) Write(p []byte) (n int, err error) {
	var x fr.Element
	for i := 0; i < len(p); i += fr.Bytes {
		x.SetBytes(p[i:min(len(p), i+fr.Bytes)])
		if err = m.write(x); err != nil {
			return
		}
	}
	n = len(p)
	return
}

func (m *MapHash) Sum(b []byte) []byte {
	mP := *m
	if _, err := mP.Write(b); err != nil {
		panic(err)
	}
	bytes := mP.state.Bytes()
	return bytes[:]
}

func (m *MapHash) Reset() {
	m.stateValid = false
}

func (m *MapHash) Size() int {
	return fr.Bytes
}

func (m *MapHash) BlockSize() int {
	return fr.Bytes
}

func (m *MapHash) write(x fr.Element) error {
	X := &x
	Y := &m.state
	if !m.stateValid {
		Y = nil
	}
	var err error
	if m.state, err = m.Map.FindPair(X, Y); err == nil {
		m.stateValid = true
	}
	return err
}

func (t *ElementTriplet) writeKey(sb *strings.Builder) {
	sb.WriteRune('"')
	sb.WriteString(t.key1.String())
	if t.key2Present {
		sb.WriteRune(',')
		sb.WriteString(t.key2.String())
	}
	sb.WriteRune('"')
}
func (m *ElementMap) UnusedEntries() []interface{} {
	unused := make([]interface{}, 0)
	for _, v := range *m {
		if !v.used {
			var vInterface interface{}
			if v.key2Present {
				vInterface = []interface{}{ElementToInterface(&v.key1), ElementToInterface(&v.key2)}
			} else {
				vInterface = ElementToInterface(&v.key1)
			}
			unused = append(unused, vInterface)
		}
	}
	return unused
}

func (m *ElementMap) sort() {
	sort.Slice(*m, func(i, j int) bool {
		return (*m)[i].CmpKey((*m)[j]) <= 0
	})
}

func (m *ElementMap) find(toFind *ElementTriplet) (fr.Element, error) {
	i := sort.Search(len(*m), func(i int) bool { return (*m)[i].CmpKey(toFind) >= 0 })

	if i < len(*m) && (*m)[i].CmpKey(toFind) == 0 {
		(*m)[i].used = true
		return (*m)[i].value, nil
	}
	var sb strings.Builder
	sb.WriteString("no value available for input ")
	toFind.writeKey(&sb)
	return fr.Element{}, fmt.Errorf(sb.String())
}

func (m *ElementMap) FindPair(x *fr.Element, y *fr.Element) (fr.Element, error) {

	toFind := ElementTriplet{
		key1:        *x,
		key2Present: y != nil,
	}

	if y != nil {
		toFind.key2 = *y
	}

	return m.find(&toFind)
}

func ToElement(i int64) *fr.Element {
	var res fr.Element
	res.SetInt64(i)
	return &res
}

type MessageCounter struct {
	startState uint64
	state      uint64
	step       uint64
}

func (m *MessageCounter) Write(p []byte) (n int, err error) {
	inputBlockSize := (len(p)-1)/fr.Bytes + 1
	m.state += uint64(inputBlockSize) * m.step
	return len(p), nil
}

func (m *MessageCounter) Sum(b []byte) []byte {
	inputBlockSize := (len(b)-1)/fr.Bytes + 1
	resI := m.state + uint64(inputBlockSize)*m.step
	var res fr.Element
	res.SetInt64(int64(resI))
	resBytes := res.Bytes()
	return resBytes[:]
}

func (m *MessageCounter) Reset() {
	m.state = m.startState
}

func (m *MessageCounter) Size() int {
	return fr.Bytes
}

func (m *MessageCounter) BlockSize() int {
	return fr.Bytes
}

func NewMessageCounter(startState, step int) hash.Hash {
	transcript := &MessageCounter{startState: uint64(startState), state: uint64(startState), step: uint64(step)}
	return transcript
}

func NewMessageCounterGenerator(startState, step int) func() hash.Hash {
	return func() hash.Hash {
		return NewMessageCounter(startState, step)
	}
}

type ListHash []fr.Element

func (h *ListHash) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (h *ListHash) Sum(b []byte) []byte {
	res := (*h)[0].Bytes()
	*h = (*h)[1:]
	return res[:]
}

func (h *ListHash) Reset() {
}

func (h *ListHash) Size() int {
	return fr.Bytes
}

func (h *ListHash) BlockSize() int {
	return fr.Bytes
}
func SetElement(z *fr.Element, value interface{}) (*fr.Element, error) {

	// TODO: Put this in element.SetString?
	switch v := value.(type) {
	case string:

		if sep := strings.Split(v, "/"); len(sep) == 2 {
			var denom fr.Element
			if _, err := z.SetString(sep[0]); err != nil {
				return nil, err
			}
			if _, err := denom.SetString(sep[1]); err != nil {
				return nil, err
			}
			denom.Inverse(&denom)
			z.Mul(z, &denom)
			return z, nil
		}

	case float64:
		asInt := int64(v)
		if float64(asInt) != v {
			return nil, fmt.Errorf("cannot currently parse float")
		}
		z.SetInt64(asInt)
		return z, nil
	}

	return z.SetInterface(value)
}

func SliceToElementSlice[T any](slice []T) ([]fr.Element, error) {
	elementSlice := make([]fr.Element, len(slice))
	for i, v := range slice {
		if _, err := SetElement(&elementSlice[i], v); err != nil {
			return nil, err
		}
	}
	return elementSlice, nil
}

func SliceEquals(a []fr.Element, b []fr.Element) error {
	if len(a) != len(b) {
		return fmt.Errorf("length mismatch %d≠%d", len(a), len(b))
	}
	for i := range a {
		if !a[i].Equal(&b[i]) {
			return fmt.Errorf("at index %d: %s ≠ %s", i, a[i].String(), b[i].String())
		}
	}
	return nil
}

func SliceSliceEquals(a [][]fr.Element, b [][]fr.Element) error {
	if len(a) != len(b) {
		return fmt.Errorf("length mismatch %d≠%d", len(a), len(b))
	}
	for i := range a {
		if err := SliceEquals(a[i], b[i]); err != nil {
			return fmt.Errorf("at index %d: %w", i, err)
		}
	}
	return nil
}

func PolynomialSliceEquals(a []polynomial.Polynomial, b []polynomial.Polynomial) error {
	if len(a) != len(b) {
		return fmt.Errorf("length mismatch %d≠%d", len(a), len(b))
	}
	for i := range a {
		if err := SliceEquals(a[i], b[i]); err != nil {
			return fmt.Errorf("at index %d: %w", i, err)
		}
	}
	return nil
}

func ElementToInterface(x *fr.Element) interface{} {
	text := x.Text(10)
	if len(text) < 10 && !strings.Contains(text, "/") {
		if i, err := strconv.Atoi(text); err != nil {
			panic(err.Error())
		} else {
			return i
		}
	}
	return text
}

func ElementSliceToInterfaceSlice(x interface{}) []interface{} {
	if x == nil {
		return nil
	}

	X := reflect.ValueOf(x)

	res := make([]interface{}, X.Len())
	for i := range res {
		xI := X.Index(i).Interface().(fr.Element)
		res[i] = ElementToInterface(&xI)
	}
	return res
}

func ElementSliceSliceToInterfaceSliceSlice(x interface{}) [][]interface{} {
	if x == nil {
		return nil
	}

	X := reflect.ValueOf(x)

	res := make([][]interface{}, X.Len())
	for i := range res {
		res[i] = ElementSliceToInterfaceSlice(X.Index(i).Interface())
	}

	return res
}
