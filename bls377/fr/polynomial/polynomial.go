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

// Code generated by gurvy DO NOT EDIT

package polynomial

import (
	"github.com/consensys/gurvy/bls377/fr"
)

// Polynomial polynomial represented by coefficients bls377 fr field.
type Polynomial []fr.Element

// Degree returns the degree of the polynomial, which is the length of Data.
func (p Polynomial) Degree() uint64 {
	res := uint64(len(p) - 1)
	return res
}

// Eval evaluates p at v
func (p Polynomial) Eval(v interface{}) interface{} {
	var res, _v fr.Element
	_v.Set(v.(*fr.Element))
	s := len(p)
	res.Set(&p[s-1])
	for i := s - 2; i >= 0; i-- {
		res.Mul(&res, &_v)
		res.Add(&res, &p[i])
	}
	return &res
}
