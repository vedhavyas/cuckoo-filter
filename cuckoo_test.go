package cuckoo

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"bytes"

	"reflect"

	"github.com/spaolacci/murmur3"
)

func Test_fingerprintOf(t *testing.T) {
	tests := []struct {
		b []byte
		r fingerprint
		h uint32
	}{
		{
			b: []byte("hello"),
			r: 26725,
			h: 19595036,
		},

		{
			b: []byte("12345"),
			r: 12594,
			h: 19870548,
		},
	}

	h := murmur3.New32WithSeed(1234)
	for _, c := range tests {
		fp := fingerprintOf(c.b)
		fph := fingerprintHash(fp, h)
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
			fp: 123,
			r:  true,
		},

		{
			i:  120,
			fp: 2345,
			r:  true,
		},

		{
			i:  156400,
			fp: 1223,
			r:  true,
		},
	}

	f := StdFilter()
	for _, c := range tests {
		r := addToBucket(&f.buckets[c.i], f.bucketSize, c.fp)
		if r != c.r {
			t.Fatalf("expected %t but got %t", c.r, r)
		}
	}
}

func TestFilter_Insert(t *testing.T) {
	tests := []struct {
		item  string
		count uint32
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
		count uint32
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

func TestFilter_EncodeDecode(t *testing.T) {
	f := StdFilter()
	data := []string{"hello", "hello, World", "This Worked"}
	for _, s := range data {
		f.Insert([]byte(s))
	}

	var b bytes.Buffer
	err := f.Encode(&b)
	if err != nil {
		t.Fatalf("unexpected error while encoding: %v", err)
	}

	df, err := Decode(&b)
	if err != nil {
		t.Fatalf("unexpected error while decoding: %v", err)
	}

	if f.count != df.count {
		t.Fatalf("count mismatch")
	}

	if f.bucketSize != df.bucketSize {
		t.Fatalf("bucketsize mismatch")
	}

	if f.totalBuckets != df.totalBuckets {
		t.Fatalf("totalbuckets mismatch")
	}

	if f.maxKicks != df.maxKicks {
		t.Fatalf("maxkicks mismatch")
	}

	if !reflect.DeepEqual(f.buckets, df.buckets) {
		t.Fatalf("buckets mismatch")
	}

	for _, s := range data {
		if !df.Lookup([]byte(s)) {
			t.Fatalf("lookup failed: %s", s)
		}
	}
}

func Test_nextPowerOf2(t *testing.T) {
	tests := []struct {
		v uint32
		e uint32
	}{
		{
			v: 1,
			e: 4,
		},

		{
			v: 2,
			e: 4,
		},

		{
			v: 3,
			e: 4,
		},

		{
			v: 4,
			e: 4,
		},

		{
			v: 100,
			e: 128,
		},
	}

	for _, c := range tests {
		g := nextPowerOf2(c.v)
		if g != c.e {
			t.Fatalf("expected %d but got %d", c.e, g)
		}
	}
}

// Benchmark tests taken from https://github.com/mtchavez/cuckoo
var filter *Filter
var okay bool

func BenchmarkCuckooNew(b *testing.B) {
	b.ReportAllocs()
	var f *Filter
	for i := 0; i < b.N; i++ {
		f = StdFilter()
	}

	filter = f
}

func BenchmarkInsert(b *testing.B) {
	var ok bool
	filter := StdFilter()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok = filter.Insert([]byte(fmt.Sprintf("item-%d", i%50000)))
	}

	okay = ok
}

func BenchmarkInsertUnique(b *testing.B) {
	var ok bool
	filter := StdFilter()
	fd, _ := os.Open("/usr/share/dict/words")
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	var wordCount int
	var totalWords int
	var values [][]byte
	for scanner.Scan() {
		word := []byte(scanner.Text())
		totalWords++

		if filter.Insert(word) {
			wordCount++
		}
		values = append(values, word)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok = filter.InsertUnique(values[i%totalWords])
	}

	okay = ok
}

func BenchmarkLookup(b *testing.B) {
	var ok bool
	filter := StdFilter()
	fd, _ := os.Open("/usr/share/dict/words")
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	var wordCount int
	var totalWords int
	var values [][]byte
	for scanner.Scan() {
		word := []byte(scanner.Text())
		totalWords++

		if filter.Insert(word) {
			wordCount++
		}
		values = append(values, word)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok = filter.Lookup(values[i%totalWords])
	}

	okay = ok
}

func BenchmarkDelete(b *testing.B) {
	var ok bool
	filter := StdFilter()
	fd, _ := os.Open("/usr/share/dict/words")
	defer fd.Close()

	scanner := bufio.NewScanner(fd)
	var wordCount int
	var totalWords int
	var values [][]byte
	for scanner.Scan() {
		word := []byte(scanner.Text())
		totalWords++

		if filter.Insert(word) {
			wordCount++
		}
		values = append(values, word)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ok = filter.Delete(values[i%totalWords])
	}

	okay = ok
}
