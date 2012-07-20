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

package step

import (
	"fmt"
	check "launchpad.net/gocheck"
	"reflect"
	"testing"
)

// Tests
func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestCreate(c *check.C) {
	_, err := New(0, 0, nil)
	c.Check(err, check.ErrorMatches, ErrZeroLength.Error())
	for _, vec := range []struct {
		start, end int
		zero       interface{}
	}{
		{1, 10, nil},
		{0, 10, nil},
		{-1, 100, nil},
		{-100, -10, nil},
		{1, 10, 0},
		{0, 10, 0},
		{-1, 100, 0},
		{-100, -10, 0},
	} {
		sv, err := New(vec.start, vec.end, vec.zero)
		c.Assert(err, check.Equals, nil)
		c.Check(sv.Start(), check.Equals, vec.start)
		c.Check(sv.End(), check.Equals, vec.end)
		c.Check(sv.Len(), check.Equals, vec.end-vec.start)
		c.Check(sv.Zero, check.DeepEquals, vec.zero)
		var at interface{}
		for i := vec.start; i < vec.end; i++ {
			at, err = sv.At(i)
			c.Check(at, check.DeepEquals, vec.zero)
			c.Check(err, check.Equals, nil)
		}
		_, err = sv.At(vec.start - 1)
		c.Check(err, check.ErrorMatches, ErrOutOfRange.Error())
		_, err = sv.At(vec.start - 1)
		c.Check(err, check.ErrorMatches, ErrOutOfRange.Error())
	}
}

func (s *S) TestSet_1(c *check.C) {
	for i, t := range []struct {
		start, end int
		zero       interface{}
		sets       []position
		expect     string
	}{
		{1, 10, 0,
			[]position{
				{1, 2},
				{2, 3},
				{3, 3},
				{4, 3},
				{5, 2},
			},
			"[1:2 2:3 5:2 6:0 10:<nil>]",
		},
		{1, 10, 0,
			[]position{
				{3, 3},
				{4, 3},
				{1, 2},
				{2, 3},
				{5, 2},
			},
			"[1:2 2:3 5:2 6:0 10:<nil>]",
		},
		{1, 10, 0,
			[]position{
				{3, 3},
				{4, 3},
				{5, 2},
				{1, 2},
				{2, 3},
				{9, 2},
			},
			"[1:2 2:3 5:2 6:0 9:2 10:<nil>]",
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		c.Check(func() { sv.Set(t.start-1, nil) }, check.Panics, ErrOutOfRange)
		c.Check(func() { sv.Set(t.end, nil) }, check.Panics, ErrOutOfRange)
		for _, v := range t.sets {
			sv.Set(v.pos, v.val)
			c.Check(sv.min.pos, check.Equals, t.start)
			c.Check(sv.max.pos, check.Equals, t.end)
			c.Check(sv.Len(), check.Equals, t.end-t.start)
		}
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
		sv.Relaxed = true
		sv.Set(t.start-1, nil)
		sv.Set(t.end, nil)
		c.Check(sv.Len(), check.Equals, t.end-t.start+2)
		for _, v := range t.sets {
			sv.Set(v.pos, t.zero)
		}
		sv.Set(t.start-1, t.zero)
		sv.Set(t.end, t.zero)
		c.Check(sv.t.Len(), check.Equals, 2)
		c.Check(sv.String(), check.Equals, fmt.Sprintf("[%d:%v %d:%v]", t.start-1, t.zero, t.end+1, nil))
	}
}

func (s *S) TestSet_2(c *check.C) {
	for i, t := range []struct {
		start, end int
		zero       interface{}
		sets       []position
		expect     string
		count      int
	}{
		{1, 2, 0,
			[]position{
				{1, 2},
				{2, 3},
				{3, 3},
				{4, 3},
				{5, 2},
				{-1, 5},
				{10, 23},
			},
			"[-1:5 0:0 1:2 2:3 5:2 6:0 10:23 11:<nil>]",
			7,
		},
		{1, 10, 0,
			[]position{
				{0, 0},
			},
			"[0:0 10:<nil>]",
			1,
		},
		{1, 10, 0,
			[]position{
				{-1, 0},
			},
			"[-1:0 10:<nil>]",
			1,
		},
		{1, 10, 0,
			[]position{
				{11, 0},
			},
			"[1:0 12:<nil>]",
			1,
		},
		{1, 10, 0,
			[]position{
				{2, 1},
				{3, 1},
				{4, 1},
				{5, 1},
				{6, 1},
				{7, 1},
				{8, 1},
				{5, 1},
			},
			"[1:0 2:1 9:0 10:<nil>]",
			3,
		},
		{1, 10, 0,
			[]position{
				{3, 1},
				{2, 1},
			},
			"[1:0 2:1 4:0 10:<nil>]",
			3,
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		sv.Relaxed = true
		for _, v := range t.sets {
			sv.Set(v.pos, v.val)
		}
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
		c.Check(sv.Count(), check.Equals, t.count)
	}
}

func (s *S) TestSetRange_1(c *check.C) {
	type posRange struct {
		start, end int
		val        interface{}
	}
	for i, t := range []struct {
		start, end int
		zero       interface{}
		sets       []posRange
		expect     string
		count      int
	}{
		{1, 10, 0,
			[]posRange{
				{1, 2, 2},
				{2, 3, 3},
				{3, 4, 3},
				{4, 5, 3},
				{5, 6, 2},
			},
			"[1:2 2:3 5:2 6:0 10:<nil>]",
			4,
		},
		{1, 10, 0,
			[]posRange{
				{3, 4, 3},
				{4, 5, 3},
				{1, 2, 2},
				{2, 3, 3},
				{5, 6, 2},
			},
			"[1:2 2:3 5:2 6:0 10:<nil>]",
			4,
		},
		{1, 10, 0,
			[]posRange{
				{3, 4, 3},
				{4, 5, 3},
				{5, 6, 2},
				{1, 2, 2},
				{2, 3, 3},
				{9, 10, 2},
			},
			"[1:2 2:3 5:2 6:0 9:2 10:<nil>]",
			5,
		},
		{1, 10, 0,
			[]posRange{
				{3, 6, 3},
				{4, 5, 1},
				{5, 7, 2},
				{1, 3, 2},
				{9, 10, 2},
			},
			"[1:2 3:3 4:1 5:2 7:0 9:2 10:<nil>]",
			6,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			"[1:3 3:0 4:1 5:0 7:2 8:0 9:4 10:<nil>]",
			7,
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		c.Check(func() { sv.SetRange(t.start-2, t.start, nil) }, check.Panics, ErrOutOfRange)
		c.Check(func() { sv.SetRange(t.end, t.end+2, nil) }, check.Panics, ErrOutOfRange)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
			c.Check(sv.min.pos, check.Equals, t.start)
			c.Check(sv.max.pos, check.Equals, t.end)
			c.Check(sv.Len(), check.Equals, t.end-t.start)
		}
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
		c.Check(sv.Count(), check.Equals, t.count)
		sv.Relaxed = true
		sv.SetRange(t.start-1, t.start, nil)
		sv.SetRange(t.end, t.end+1, nil)
		c.Check(sv.Len(), check.Equals, t.end-t.start+2)
		sv.SetRange(t.start-1, t.end+1, t.zero)
		c.Check(sv.t.Len(), check.Equals, 2)
		c.Check(sv.String(), check.Equals, fmt.Sprintf("[%d:%v %d:%v]", t.start-1, t.zero, t.end+1, nil))
	}
}

func (s *S) TestSetRange_2(c *check.C) {
	sv, _ := New(0, 1, nil)
	c.Check(func() { sv.SetRange(1, 0, nil) }, check.Panics, ErrInvertedRange)
	type posRange struct {
		start, end int
		val        interface{}
	}
	for i, t := range []struct {
		start, end int
		zero       interface{}
		sets       []posRange
		expect     string
	}{
		{1, 10, 0,
			[]posRange{
				{1, 2, 2},
				{2, 3, 3},
				{3, 4, 3},
				{4, 5, 3},
				{5, 6, 2},
				{-10, -1, 4},
				{23, 35, 10},
			},
			"[-10:4 -1:0 1:2 2:3 5:2 6:0 23:10 35:<nil>]",
		},
		{1, 2, 0,
			[]posRange{
				{1, 1, 2},
			},
			"[1:0 2:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{-10, 1, 0},
			},
			"[-10:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{-10, 1, 1},
			},
			"[-10:1 1:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{-10, 0, 1},
			},
			"[-10:1 0:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{-10, 0, 0},
			},
			"[-10:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{10, 20, 0},
			},
			"[1:0 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{10, 20, 1},
			},
			"[1:0 10:1 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{11, 20, 0},
			},
			"[1:0 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{11, 20, 1},
			},
			"[1:0 11:1 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{1, 10, 1},
				{11, 20, 1},
			},
			"[1:1 10:0 11:1 20:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{2, 5, 0},
			},
			"[1:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 6, 1},
				{2, 5, 0},
			},
			"[1:0 5:1 6:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 1},
				{5, 7, 2},
				{3, 5, 1},
			},
			"[1:1 5:2 7:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 1},
				{5, 7, 2},
				{3, 5, 2},
			},
			"[1:1 3:2 7:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{2, 6, 0},
			},
			"[1:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 6, 1},
				{2, 5, 0},
			},
			"[1:0 5:1 6:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{2, 5, 2},
			},
			"[1:0 2:2 5:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{3, 5, 2},
			},
			"[1:0 2:1 3:2 5:0 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{2, 5, 1},
				{3, 5, 0},
			},
			"[1:0 2:1 3:0 10:<nil>]",
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		sv.Relaxed = true
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
	}
}

func (s *S) TestStepAt(c *check.C) {
	type posRange struct {
		start, end int
		val        interface{}
	}
	t := struct {
		start, end int
		zero       interface{}
		sets       []posRange
		expect     string
	}{1, 10, 0,
		[]posRange{
			{1, 3, 3},
			{4, 5, 1},
			{7, 8, 2},
			{9, 10, 4},
		},
		"[1:3 3:0 4:1 5:0 7:2 8:0 9:4 10:<nil>]",
	}

	sv, err := New(t.start, t.end, t.zero)
	c.Assert(err, check.Equals, nil)
	for _, v := range t.sets {
		sv.SetRange(v.start, v.end, v.val)
	}
	c.Check(sv.String(), check.Equals, t.expect)
	for i, v := range t.sets {
		for j := v.start; j < v.end; j++ {
			at, st, en, err := sv.StepAt(v.start)
			c.Check(err, check.Equals, nil)
			c.Check(at, check.DeepEquals, v.val)
			c.Check(st, check.Equals, v.start)
			c.Check(en, check.Equals, v.end)
		}
		at, st, en, err := sv.StepAt(v.end)
		if v.end < sv.End() {
			c.Check(err, check.Equals, nil)
			c.Check(at, check.DeepEquals, sv.Zero)
			c.Check(st, check.Equals, v.end)
			c.Check(en, check.Equals, t.sets[i+1].start)
		} else {
			c.Check(err, check.ErrorMatches, ErrOutOfRange.Error())
		}
	}
	_, _, _, err = sv.StepAt(t.start - 1)
	c.Check(err, check.ErrorMatches, ErrOutOfRange.Error())
}

func (s *S) TestDo(c *check.C) {
	var data interface{}
	type posRange struct {
		start, end int
		val        interface{}
	}
	for i, t := range []struct {
		start, end int
		zero       interface{}
		sets       []posRange
		setup      func()
		fn         func(start, end int, v interface{})
		expect     interface{}
	}{
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func() { data = []int(nil) },
			func(start, end int, vi interface{}) {
				sl := data.([]int)
				v := vi.(int)
				for i := start; i < end; i++ {
					sl = append(sl, v)
				}
				data = sl
			},
			[]int{3, 3, 0, 1, 0, 0, 2, 0, 4},
		},
	} {
		t.setup()
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		sv.Do(t.fn)
		c.Check(data, check.DeepEquals, t.expect, check.Commentf("subtest %d", i))
		c.Check(reflect.ValueOf(data).Len(), check.Equals, sv.Len())
	}
}

func (s *S) TestDoRange(c *check.C) {
	var data interface{}
	type posRange struct {
		start, end int
		val        interface{}
	}
	for i, t := range []struct {
		start, end int
		zero       interface{}
		sets       []posRange
		setup      func()
		fn         func(start, end int, v interface{})
		from, to   int
		expect     interface{}
		err        error
	}{
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func() { data = []int(nil) },
			func(start, end int, vi interface{}) {
				sl := data.([]int)
				v := vi.(int)
				for i := start; i < end; i++ {
					sl = append(sl, v)
				}
				data = sl
			},
			2, 8,
			[]int{3, 0, 1, 0, 0, 2},
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func() { data = []int(nil) },
			func(_, _ int, _ interface{}) {},
			-2, -1,
			[]int(nil),
			ErrOutOfRange,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func() { data = []int(nil) },
			func(_, _ int, _ interface{}) {},
			10, 1,
			[]int(nil),
			ErrInvertedRange,
		},
	} {
		t.setup()
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		c.Check(sv.DoRange(t.fn, t.from, t.to), check.DeepEquals, t.err)
		c.Check(data, check.DeepEquals, t.expect, check.Commentf("subtest %d", i))
		if t.from <= t.to && t.from < sv.End() && t.to > sv.Start() {
			c.Check(reflect.ValueOf(data).Len(), check.Equals, t.to-t.from)
		}
	}
}

func (s *S) TestApply(c *check.C) {
	type posRange struct {
		start, end int
		val        interface{}
	}
	for i, t := range []struct {
		start, end int
		zero       interface{}
		sets       []posRange
		mutate     func(interface{}) interface{}
		expect     string
	}{
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			IncInt,
			"[1:4 3:1 4:2 5:1 7:3 8:1 9:5 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			DecInt,
			"[1:2 3:-1 4:0 5:-1 7:1 8:-1 9:3 10:<nil>]",
		},
		{1, 10, 0.,
			[]posRange{
				{1, 3, 3.},
				{4, 5, 1.},
				{7, 8, 2.},
				{9, 10, 4.},
			},
			IncFloat,
			"[1:4 3:1 4:2 5:1 7:3 8:1 9:5 10:<nil>]",
		},
		{1, 10, 0.,
			[]posRange{
				{1, 3, 3.},
				{4, 5, 1.},
				{7, 8, 2.},
				{9, 10, 4.},
			},
			DecFloat,
			"[1:2 3:-1 4:0 5:-1 7:1 8:-1 9:3 10:<nil>]",
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func(_ interface{}) interface{} { return 0 },
			"[1:0 10:<nil>]",
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		sv.Apply(t.mutate)
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
	}
}

func (s *S) TestMutateRange(c *check.C) {
	type posRange struct {
		start, end int
		val        interface{}
	}
	for i, t := range []struct {
		start, end int
		zero       interface{}
		sets       []posRange
		mutate     func(interface{}) interface{}
		from, to   int
		expect     string
		err        error
	}{
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			IncInt,
			2, 8,
			"[1:3 2:4 3:1 4:2 5:1 7:3 8:0 9:4 10:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{7, 8, 2},
				{9, 10, 4},
			},
			IncInt,
			4, 6,
			"[1:3 3:0 4:1 6:0 7:2 8:0 9:4 10:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{7, 8, 1},
				{9, 10, 4},
			},
			IncInt,
			4, 7,
			"[1:3 3:0 4:1 8:0 9:4 10:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			func(_ interface{}) interface{} { return 0 },
			2, 8,
			"[1:3 2:0 9:4 10:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{7, 8, 1},
				{9, 10, 4},
			},
			IncInt,
			4, 8,
			"[1:3 3:0 4:1 7:2 8:0 9:4 10:<nil>]",
			nil,
		},
		{1, 20, 0,
			[]posRange{
				{5, 10, 1},
				{10, 15, 2},
				{15, 20, 3},
			},
			func(v interface{}) interface{} {
				if v.(int) == 3 {
					return 1
				}
				return v
			},
			8, 18,
			"[1:0 5:1 10:2 15:1 18:3 20:<nil>]",
			nil,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			IncInt,
			-1, 0,
			"[1:3 3:0 4:1 5:0 7:2 8:0 9:4 10:<nil>]",
			ErrOutOfRange,
		},
		{1, 10, 0,
			[]posRange{
				{1, 3, 3},
				{4, 5, 1},
				{7, 8, 2},
				{9, 10, 4},
			},
			IncInt,
			10, 1,
			"[1:3 3:0 4:1 5:0 7:2 8:0 9:4 10:<nil>]",
			ErrInvertedRange,
		},
	} {
		sv, err := New(t.start, t.end, t.zero)
		c.Assert(err, check.Equals, nil)
		for _, v := range t.sets {
			sv.SetRange(v.start, v.end, v.val)
		}
		c.Check(sv.ApplyRange(t.mutate, t.from, t.to), check.DeepEquals, t.err)
		c.Check(sv.String(), check.Equals, t.expect, check.Commentf("subtest %d", i))
	}
}
