// Copyright ©2012 Dan Kortschak <dan.kortschak@adelaide.edu.au>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

// Package step implements a step vector type.
//
// A step vector can be used to represent high volume data that would be
// efficiently stored by run-length encoding.
package step

import (
	"code.google.com/p/biogo.llrb"
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrOutOfRange    = errors.New("step: index out of range")
	ErrInvertedRange = errors.New("step: inverted range")
	ErrZeroLength    = errors.New("step: attempt to create zero length vector")
)

type (
	position struct {
		pos int
		val interface{} // This must be able to be compared by reflect.DeepEqual
	}
	query int
	upper int
)

func (p *position) Compare(c llrb.Comparable) int {
	return p.pos - c.(*position).pos
}
func (q query) Compare(c llrb.Comparable) (d int) {
	switch c := c.(type) {
	case *position:
		d = int(q) - c.pos
	case query:
		d = int(q) - int(c)
	}
	return
}
func (q upper) Compare(c llrb.Comparable) (d int) {
	d = int(q) - c.(*position).pos
	if d == 0 {
		d = 1
	}
	return
}

// A Vector is type that support the storage of array type data in a run-length
// encoding format.
type Vector struct {
	Zero     interface{} // Ground state for the step vector.
	Relaxed  bool        // If true, dynamic vector resize is allowed.
	t        *llrb.Tree
	min, max *position
}

// New returns a new Vector with the extent defined by start and end,
// and the ground state defined by zero. The Vector's extent is mutable
// if the Relaxed field is set to true. If a zero length vector is requested
// an error is returned.
func New(start, end int, zero interface{}) (v *Vector, err error) {
	if start >= end {
		return nil, ErrZeroLength
	}
	v = &Vector{
		Zero: zero,
		t:    &llrb.Tree{},
		min: &position{
			pos: start,
			val: zero,
		},
		max: &position{
			pos: end,
			val: nil,
		},
	}
	v.t.Insert(v.min)
	v.t.Insert(v.max)

	return
}

// Start returns the index of minimum position of the Vector.
func (self *Vector) Start() int { return self.min.pos }

// End returns the index of lowest position beyond the end of the Vector.
func (self *Vector) End() int { return self.max.pos }

// Len returns the length of the represented data array, that is the distance
// between the start and end of the vector.
func (self *Vector) Len() int { return self.End() - self.Start() }

// Count returns the number of steps represented in the vector.
func (self *Vector) Count() int { return self.t.Len() - 1 }

// At returns the value of the vector at position i. If i is outside the extent
// of the vector an error is returned.
func (self *Vector) At(i int) (v interface{}, err error) {
	if i < self.Start() || i >= self.End() {
		return nil, ErrOutOfRange
	}
	st := self.t.Floor(query(i)).(*position)
	return st.val, nil
}

// StepAt returns the value and range of the step at i, where start <= i < end.
// If i is outside the extent of the vector, an error is returned.
func (self *Vector) StepAt(i int) (v interface{}, start, end int, err error) {
	if i < self.Start() || i >= self.End() {
		return nil, 0, 0, ErrOutOfRange
	}
	lo := self.t.Floor(query(i)).(*position)
	hi := self.t.Ceil(upper(i)).(*position)
	return lo.val, lo.pos, hi.pos, nil
}

// Set sets the value of position i to v.
// The underlying type of v must be comparable by reflect.DeepEqual.
func (self *Vector) Set(i int, v interface{}) {
	if i < self.min.pos || self.max.pos <= i {
		if !self.Relaxed {
			panic(ErrOutOfRange)
		}

		if i < self.min.pos {
			if i == self.min.pos-1 {
				if reflect.DeepEqual(v, self.min.val) {
					self.min.pos--
				} else {
					self.min = &position{pos: i, val: v}
					self.t.Insert(self.min)
				}
			} else {
				if reflect.DeepEqual(self.min.val, self.Zero) {
					self.min.pos = i + 1
				} else {
					self.min = &position{pos: i + 1, val: self.Zero}
					self.t.Insert(self.min)
				}
				if reflect.DeepEqual(v, self.Zero) {
					self.min.pos--
				} else {
					self.min = &position{pos: i, val: v}
					self.t.Insert(self.min)
				}
			}
		} else if i >= self.max.pos {
			if i == self.max.pos {
				self.max.pos++
				prev := self.t.Floor(query(i)).(*position)
				if !reflect.DeepEqual(v, prev.val) {
					self.t.Insert(&position{pos: i, val: v})
				}
			} else {
				mpos := self.max.pos
				self.max.pos = i + 1
				prev := self.t.Floor(query(i)).(*position)
				if !reflect.DeepEqual(prev.val, self.Zero) {
					self.t.Insert(&position{pos: mpos, val: self.Zero})
				}
				if !reflect.DeepEqual(v, self.Zero) {
					self.t.Insert(&position{pos: i, val: v})
				}
			}
		}
		return
	}

	lo := self.t.Floor(query(i)).(*position)
	if reflect.DeepEqual(v, lo.val) {
		return
	}
	hi := self.t.Ceil(upper(i)).(*position)

	if lo.pos == i {
		if hi.pos == i+1 {
			if reflect.DeepEqual(v, hi.val) {
				self.t.Delete(query(i))
				hi.pos--
			} else {
				lo.val = v
			}
			if i > self.min.pos {
				prev := self.t.Floor(query(i - 1)).(*position)
				if reflect.DeepEqual(v, prev.val) {
					self.t.Delete(query(i))
				}
			}
		} else {
			lo.pos = i + 1
			prev := self.t.Floor(query(i))
			if prev == nil {
				self.min = &position{pos: i, val: v}
				self.t.Insert(self.min)
			} else if !reflect.DeepEqual(v, prev.(*position).val) {
				self.t.Insert(&position{pos: i, val: v})
			}
		}
	} else {
		if hi.pos == i+1 {
			if reflect.DeepEqual(v, hi.val) {
				hi.pos--
			} else {
				self.t.Insert(&position{pos: i, val: v})
			}
		} else {
			self.t.Insert(&position{pos: i, val: v})
			self.t.Insert(&position{pos: i + 1, val: lo.val})
		}
	}
}

// SetRange sets the value of positions [start, end) to v.
// The underlying type of v must be comparable by reflect.DeepEqual.
func (self *Vector) SetRange(start, end int, v interface{}) {
	l := end - start
	switch {
	case l == 0:
		return
	case l == 1:
		self.Set(start, v)
		return
	case l < 0:
		panic(ErrInvertedRange)
	}

	if end <= self.min.pos || self.max.pos <= start {
		if !self.Relaxed {
			panic(ErrOutOfRange)
		}

		if end <= self.min.pos {
			if end == self.min.pos {
				if reflect.DeepEqual(v, self.min.val) {
					self.min.pos -= l
				} else {
					self.min = &position{pos: start, val: v}
					self.t.Insert(self.min)
				}
			} else {
				if reflect.DeepEqual(self.min.val, self.Zero) {
					self.min.pos = end
				} else {
					self.min = &position{pos: end, val: self.Zero}
					self.t.Insert(self.min)
				}
				if reflect.DeepEqual(v, self.Zero) {
					self.min.pos -= l
				} else {
					self.min = &position{pos: start, val: v}
					self.t.Insert(self.min)
				}
			}
		} else if start >= self.max.pos {
			if start == self.max.pos {
				self.max.pos += l
				prev := self.t.Floor(query(start)).(*position)
				if !reflect.DeepEqual(v, prev.val) {
					self.t.Insert(&position{pos: start, val: v})
				}
			} else {
				mpos := self.max.pos
				self.max.pos = end
				prev := self.t.Floor(query(start)).(*position)
				if !reflect.DeepEqual(prev.val, self.Zero) {
					self.t.Insert(&position{pos: mpos, val: self.Zero})
				}
				if !reflect.DeepEqual(v, self.Zero) {
					self.t.Insert(&position{pos: start, val: v})
				}
			}
		}
		return
	}

	delQ := []llrb.Comparable{}
	self.t.DoRange(func(c llrb.Comparable) (done bool) {
		delQ = append(delQ, c)
		return
	}, query(start), query(end))
	for _, p := range delQ {
		self.t.Delete(p)
	}

	var la, lo *position
	if len(delQ) > 0 {
		lo = delQ[0].(*position)
		la = delQ[len(delQ)-1].(*position)
	} else {
		lo = self.t.Floor(query(start)).(*position)
		la = &position{}
		*la = *lo
	}

	hi := self.t.Ceil(query(end)).(*position)
	if start == lo.pos {
		var prevSame bool
		prev := self.t.Floor(query(start - 1))
		if prev != nil {
			prevSame = reflect.DeepEqual(v, prev.(*position).val)
		}
		hiSame := reflect.DeepEqual(v, hi.val)
		if hi.pos == end {
			switch {
			case hiSame && prevSame:
				self.t.Delete(hi)
			case prevSame:
				return
			case hiSame:
				hi.pos = start
			default:
				if prev == nil {
					self.min = &position{pos: start, val: v}
					self.t.Insert(self.min)
				} else {
					self.t.Insert(&position{pos: start, val: v})
				}
			}
		} else {
			la.pos = end
			if !reflect.DeepEqual(v, la.val) {
				self.t.Insert(la)
			}
			if prev == nil {
				self.min = &position{pos: start, val: v}
				self.t.Insert(self.min)
			} else if !prevSame {
				self.t.Insert(&position{pos: start, val: v})
			}
		}
	} else {
		if hi.pos == end {
			if reflect.DeepEqual(v, hi.val) {
				hi.pos = start
			} else {
				self.t.Insert(&position{pos: start, val: v})
			}
		} else {
			self.t.Insert(&position{pos: start, val: v})
			la.pos = end
			if !reflect.DeepEqual(v, la.val) {
				self.t.Insert(la)
			}
		}
	}
}

// Do performs the function fn on steps stored in the Vector in ascending sort order
// of start position. fn is passed the start, end and value of the step.
func (self *Vector) Do(fn func(start, end int, v interface{})) {
	var (
		la  *position
		min = self.min.pos
	)

	self.t.Do(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		if p.pos != min {
			fn(la.pos, p.pos, la.val)
		}
		la = p
		return
	})
}

// Do performs the function fn on steps stored in the Vector over the range [from, to)
// in ascending sort order of start position. fn is passed the start, end and value of
// the step.
func (self *Vector) DoRange(fn func(start, end int, v interface{}), from, to int) (err error) {
	if to < from {
		return ErrInvertedRange
	}
	var (
		la  *position
		min = self.min.pos
		max = self.max.pos
	)
	if to <= min || from >= max {
		return ErrOutOfRange
	}

	v, _, end, _ := self.StepAt(from)
	fn(from, end, v)
	self.t.DoRange(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		if p.pos != end {
			fn(la.pos, p.pos, la.val)
		}
		la = p
		return
	}, query(end), query(to))
	if to > la.pos {
		fn(la.pos, to, la.val)
	}

	return
}

// Convenience mutator functions. Mutator functions are used by Apply and ApplyRange
// to alter step values in a value-dependent manner. These mutators assume the stored
// type matches the function and will panic is this is not true.
var (
	IncInt   = incInt   // Increment an int value.
	DecInt   = decInt   // Decrement an int value.
	IncFloat = incFloat // Increment a float64 value.
	DecFloat = decFloat // Decrement a float64 value.
)

func incInt(v interface{}) interface{}   { return v.(int) + 1 }
func decInt(v interface{}) interface{}   { return v.(int) - 1 }
func incFloat(v interface{}) interface{} { return v.(float64) + 1 }
func decFloat(v interface{}) interface{} { return v.(float64) - 1 }

// Apply applies the mutator function m to steps stored in the Vector in ascending sort order
// of start position. Redundant steps resulting from changes in step values are erased.
func (self *Vector) Apply(m func(interface{}) interface{}) {
	var (
		la   interface{}
		min  = self.min.pos
		max  = self.max.pos
		delQ []query
	)

	self.t.Do(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		if p.pos == max {
			return true
		}
		p.val = m(p.val)
		if p.pos != min && p.pos != max && reflect.DeepEqual(p.val, la) {
			delQ = append(delQ, query(p.pos))
		}
		la = p.val
		return
	})

	for _, d := range delQ {
		self.t.Delete(d)
	}
}

// Apply applies the mutator function m to steps stored in the Vector in over the range
// [from, to) in ascending sort order of start position. Redundant steps resulting from
// changes in step values are erased.
func (self *Vector) ApplyRange(m func(interface{}) interface{}, from, to int) (err error) {
	if to < from {
		return ErrInvertedRange
	}
	var (
		la   interface{}
		old  position
		min  = self.min.pos
		max  = self.max.pos
		delQ []query
	)
	if to <= min || from >= max {
		return ErrOutOfRange
	}

	var end int
	old.val, old.pos, end, _ = self.StepAt(from)
	la = old.val
	la = m(la)
	if to <= end {
		self.SetRange(from, to, la)
		return
	}
	if !reflect.DeepEqual(la, old.val) {
		self.t.Insert(&position{from, la})
	}
	self.t.DoRange(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		if p.pos == max {
			return true
		}
		old = *p // Needed for fix-up of last step if to is not at a step boundary.
		p.val = m(p.val)
		if p.pos != min && reflect.DeepEqual(p.val, la) {
			delQ = append(delQ, query(p.pos))
		}
		la = p.val
		return
	}, query(end), query(to))

	if to < max {
		p := self.t.Ceil(query(to)).(*position)
		if p.pos > to && !reflect.DeepEqual(p.val, old.val) {
			self.t.Insert(&position{pos: to, val: old.val})
		} else if reflect.DeepEqual(p.val, la) {
			delQ = append(delQ, query(p.pos))
		}
	}

	for _, d := range delQ {
		self.t.Delete(d)
	}

	return
}

// String returns a string representation a Vector, displaying step start
// positions and values. The last step indicates the end of the vector and
// always has an associated value of nil.
func (self *Vector) String() string {
	sb := []string(nil)
	self.t.Do(func(c llrb.Comparable) (done bool) {
		p := c.(*position)
		sb = append(sb, fmt.Sprintf("%d:%v", p.pos, p.val))
		return
	})
	return fmt.Sprintf("%v", sb)
}
