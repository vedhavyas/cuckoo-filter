package cuckoo

import (
	"bytes"
	"hash"
	"math"
	"sync"

	"math/rand"

	"fmt"

	"github.com/spaolacci/murmur3"
)

const (
	DefaultBucketSize        = 4
	DefaultTotalBuckets      = 250000
	DefaultMaxKick           = 500
	DefaultFaultPositiveRate = 3
	seed                     = 59053
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
	totalBuckets      uint64
	fingerprintSize   int
	hash              hash.Hash64
	maxKicks          int

	// protects above fields
	mu *sync.RWMutex
}

// calculateFingerprintSize calculates the fingerprint size from
// e - false positive percent and b - bucket size
func calculateFingerprintSizeInBytes(e float64, b uint64) int {
	return int(math.Ceil((math.Log(float64(100)/e) + math.Log(2*float64(b))) / 8))
}

func hashOf(x []byte, hash hash.Hash64) uint64 {
	hash.Reset()
	hash.Write(x)
	return hash.Sum64()
}

// fingerprintOf returns the fingerprint of x with size using hash
func fingerprintOf(x []byte, fpSize int, hash hash.Hash64) (fp fingerprint, fph uint64) {
	hash.Reset()
	hash.Write(x)
	fp = make(fingerprint, fpSize)
	copy(fp, hash.Sum(nil))
	return fp, hashOf(fp, hash)
}

// indicesOf returns the indices of item x using given hash
func indicesOf(x []byte, fph, totalBuckets uint64, hash hash.Hash64) (i1, i2 uint64) {
	hash.Reset()
	hash.Write(x)
	i1 = hash.Sum64() % totalBuckets
	i2 = (i1 ^ fph) % totalBuckets
	return i1, i2
}

func initBuckets(totalBuckets uint64, bucketSize int) []bucket {
	buckets := make([]bucket, totalBuckets, totalBuckets)
	for i := range buckets {
		buckets[i] = make([]fingerprint, bucketSize, bucketSize)
	}

	return buckets
}

// DefaultFilter returns filter with default values
func DefaultFilter() *Filter {
	return &Filter{
		buckets:           initBuckets(DefaultTotalBuckets, DefaultBucketSize),
		falsePositiveRate: DefaultFaultPositiveRate,
		bucketSize:        DefaultBucketSize,
		totalBuckets:      DefaultTotalBuckets,
		fingerprintSize:   calculateFingerprintSizeInBytes(DefaultFaultPositiveRate, DefaultBucketSize),
		hash:              murmur3.New64WithSeed(seed),
		maxKicks:          DefaultMaxKick,
	}
}

// addToBucket will add fp to the bucket i in filter
func addToBucket(b bucket, fp fingerprint) bool {
	for j := range b {
		if !bytes.Equal(b[j], emptyFingerprint) {
			continue
		}

		b[j] = fp
		return true
	}

	return false
}

func replaceItem(f *Filter, i uint64, fp fingerprint) (j uint64, rfp fingerprint) {
	k := rand.Intn(len(f.buckets[i]))
	rfp = f.buckets[i][k]
	f.buckets[i][k] = fp
	rfph := hashOf(rfp, f.hash)
	j = (i ^ rfph) % f.totalBuckets
	return j, rfp
}

func insert(f *Filter, x []byte) error {
	fp, fph := fingerprintOf(x, f.fingerprintSize, f.hash)
	i1, i2 := indicesOf(x, fph, f.totalBuckets, f.hash)

	if addToBucket(f.buckets[i1], fp) || addToBucket(f.buckets[i2], fp) {
		return nil
	}

	is := []uint64{i1, i2}
	i1 = is[rand.Intn(len(is))]
	for k := 0; k < f.maxKicks; k++ {
		i1, fp = replaceItem(f, i1, fp)
		if addToBucket(f.buckets[i1], fp) {
			return nil
		}
	}

	return fmt.Errorf("reached max kicks: %d", f.maxKicks)
}

//func exists(f *Filter, x []byte) bool {
//	fp, fph := fingerprintOf(x, f.fingerprintSize, f.hash)
//	i1, i2 := indicesOf(x, fph, f.totalBuckets, f.hash)
//
//	if addToBucket(f, i1, fp) || addToBucket(f, i2, fp) {
//		return tr
//	}
//}

// Insert inserts the item to the filter
// returns error of filter is full
func (f *Filter) Insert(x []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	return insert(f, x)
}
