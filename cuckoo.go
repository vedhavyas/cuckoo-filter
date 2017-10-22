package cuckoo

import (
	"hash"
	"math"
	"sync"
)

const (
	DefaultBucketSize   = 4
	DefaultTotalBuckets = 1073741824
	DefaultTotalItems   = 4294967296
	seed                = 59053
)

// emptyFingerprint represents an empty fingerprint
var emptyFingerprint []byte

// fingerprint of the item
type fingerprint []byte

// bucket with b fingerprint per bucket
type bucket []fingerprint

// Filter holds the bucket
type Filter struct {
	count             uint64
	buckets           []bucket
	falsePositiveRate float64
	bucketSize        uint64
	fingerprintSize   int
	hash              hash.Hash64

	// protects above fields
	mu *sync.RWMutex
}

// calculateFingerprintSize calculates the fingerprint size from
// e - false positive percent and b - bucket size
func calculateFingerprintSizeInBytes(e float64, b uint64) int {
	return int(math.Ceil((math.Log(float64(100)/e) + math.Log(2*float64(b))) / 8))
}

// fingerprintOf returns the fingerprint of x with size using hash
func fingerprintOf(x []byte, size int, hash hash.Hash64) (fp fingerprint, fph uint64) {
	hash.Reset()
	hash.Write(x)
	fp = make(fingerprint, size)
	copy(fp, hash.Sum(nil))
	return fp, hash.Sum64()
}

// indicesOf returns the indices of item x using given hash
func indicesOf(x []byte, fph uint64, hash hash.Hash64) (i1, i2 uint64) {
	hash.Reset()
	hash.Write(x)
	i1 = hash.Sum64()
	i2 = i1 ^ fph
	return i1, i2
}
