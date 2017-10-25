/*
Benchmark tests taken from https://github.com/mtchavez/cuckoo
*/
package cuckoo

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

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
