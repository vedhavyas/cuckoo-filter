package cuckoo

import (
	"strconv"
	"testing"

	"github.com/spaolacci/murmur3"
)

func Test_calculateFingerprintSize(t *testing.T) {
	tests := []struct {
		e float64
		b uint64
		s int
	}{
		{
			e: 3,
			b: 4,
			s: 1,
		},

		{
			e: 2,
			b: 8,
			s: 1,
		},
	}

	for _, c := range tests {
		r := calculateFingerprintSizeInBytes(c.e, c.b)
		if r != c.s {
			t.Fatalf("expected %d size but got %d", c.s, r)
		}
	}
}

func Test_fingerprintOf(t *testing.T) {
	tests := []struct {
		b []byte
		s int
		r fingerprint
		h uint64
	}{
		{
			b: []byte("hello"),
			s: 3,
			r: 36959,
			h: 12338335298928788839,
		},

		{
			b: []byte(strconv.Itoa(12345)),
			s: 5,
			r: 62830,
			h: 2238787319979236306,
		},
	}

	h := murmur3.New64WithSeed(1234)
	for _, c := range tests {
		fp, fph := fingerprintOf(c.b, c.s, h)
		if c.r != fp {
			t.Fatalf("expected %v bytes but got %v", c.r, fp)
		}

		if fph != c.h {
			t.Fatalf("expected %d hash but got %d", c.h, fph)
		}
	}
}

func Test_addToBucket(t *testing.T) {
	tests := []struct {
		i  uint64
		fp fingerprint
		r  bool
	}{
		{
			i:  100,
			fp: 36959,
			r:  true,
		},

		{
			i:  120,
			fp: 1232,
			r:  true,
		},

		{
			i:  156400,
			fp: 7626,
			r:  true,
		},
	}

	f := StdFilter()
	for _, c := range tests {
		r := addToBucket(f.buckets[c.i], c.fp)
		if r != c.r {
			t.Fatalf("expected %t but got %t", c.r, r)
		}
	}
}

func TestFilter_Insert(t *testing.T) {
	tests := []struct {
		item  string
		count uint64
	}{
		{
			item:  "hello",
			count: 1,
		},

		{
			item:  "hello, World",
			count: 2,
		},

		{
			item:  "hello, World",
			count: 3,
		},
	}

	f := StdFilter()
	for _, c := range tests {
		f.Insert([]byte(c.item))
		if f.Count() != c.count {
			t.Fatalf("expected %d count but got %d", c.count, f.Count())
		}
	}
}

func TestFilter_Exists(t *testing.T) {
	f := StdFilter()
	for _, s := range []string{"hello", "hello, World", "This Worked"} {
		f.Insert([]byte(s))
	}

	tests := []struct {
		item  string
		exist bool
	}{
		{
			item:  "hello",
			exist: true,
		},

		{
			item:  "hello, World",
			exist: true,
		},

		{
			item: "This is test11",
		},

		{
			item:  "This Worked",
			exist: true,
		},
	}

	for _, c := range tests {
		ok := f.Lookup([]byte(c.item))
		if ok != c.exist {
			t.Fatalf("extected %s item to give %t but gave %t", c.item, c.exist, ok)
		}
	}
}

func TestFilter_Delete(t *testing.T) {
	f := StdFilter()
	for _, s := range []string{"hello", "hello, World", "This Worked"} {
		f.Insert([]byte(s))
	}

	tests := []struct {
		item  string
		ok    bool
		count uint64
	}{
		{
			item:  "hello",
			ok:    true,
			count: 2,
		},

		{
			item:  "hello, World",
			ok:    true,
			count: 1,
		},

		{
			item:  "This is test",
			count: 1,
		},

		{
			item:  "This Worked",
			ok:    true,
			count: 0,
		},
	}

	for _, c := range tests {
		ok := f.Delete([]byte(c.item))
		if ok != c.ok {
			t.Fatalf("extected %s item to give %t but gave %t", c.item, c.ok, ok)
		}

		if f.Count() != c.count {
			t.Fatalf("expected %d count but got %d", c.count, f.Count())
		}
	}
}
