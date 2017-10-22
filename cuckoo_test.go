package cuckoo

import (
	"bytes"
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
		r []byte
		h uint64
	}{
		{
			b: []byte("hello"),
			s: 3,
			r: []byte{144, 95, 145},
			h: 10403193130508565092,
		},

		{
			b: []byte(strconv.Itoa(12345)),
			s: 5,
			r: []byte{245, 110, 76, 206, 27},
			h: 17685157234837869622,
		},
	}

	h := murmur3.New64WithSeed(1234)
	for _, c := range tests {
		fp, fph := fingerprintOf(c.b, c.s, h)
		if !bytes.Equal(fp, c.r) {
			t.Fatalf("expected %v bytes but got %v", c.r, fp)
		}

		if fph != c.h {
			t.Fatalf("expected %d hash but got %d", c.h, fph)
		}
	}
}
