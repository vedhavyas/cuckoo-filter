package cuckoo

import (
	"bytes"
	"hash"
	"math"
	"math/rand"
	"sync"

	"github.com/spaolacci/murmur3"
)

const (
	defaultBucketSize        = 4
	defaultTotalBuckets      = 250000
	defaultMaxKicks          = 500
	defaultFaultPositiveRate = 3
	seed                     = 59053
)

// emptyFingerprint represents an empty fingerprint
var emptyFingerprint []byte

// fingerprint of the item
type fingerprint []byte

// bucket with b fingerprints per bucket
type bucket []fingerprint

// Filter is the cuckoo-filter
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

// hashOf returns the 64-bit hash
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

// initBuckets initialises the buckets
func initBuckets(totalBuckets uint64, bucketSize int) []bucket {
	buckets := make([]bucket, totalBuckets, totalBuckets)
	for i := range buckets {
		buckets[i] = make([]fingerprint, bucketSize, bucketSize)
	}

	return buckets
}

// StdFilter returns Standard Cuckoo-Filter
func StdFilter() *Filter {
	return &Filter{
		buckets:           initBuckets(defaultTotalBuckets, defaultBucketSize),
		falsePositiveRate: defaultFaultPositiveRate,
		bucketSize:        defaultBucketSize,
		totalBuckets:      defaultTotalBuckets,
		fingerprintSize:   calculateFingerprintSizeInBytes(defaultFaultPositiveRate, defaultBucketSize),
		hash:              murmur3.New64WithSeed(seed),
		maxKicks:          defaultMaxKicks,
		mu:                &sync.RWMutex{},
	}
}

// deleteFrom deletes fingerprint from bucket if exists
func deleteFrom(b bucket, fp fingerprint) bool {
	for i := range b {
		if !bytes.Equal(b[i], fp) {
			continue
		}

		b[i] = emptyFingerprint
		return true
	}

	return false
}

// containsIn returns if the given fingerprint exists in bucket
func containsIn(b bucket, fp fingerprint) bool {
	for i := range b {
		if bytes.Equal(b[i], fp) {
			return true
		}
	}

	return false
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

// replaceItem replaces fingerprint from i and returns the alternate index for kicked fingerprint
func replaceItem(f *Filter, i uint64, fp fingerprint) (j uint64, rfp fingerprint) {
	k := rand.Intn(len(f.buckets[i]))
	rfp = f.buckets[i][k]
	f.buckets[i][k] = fp
	rfph := hashOf(rfp, f.hash)
	j = (i ^ rfph) % f.totalBuckets
	return j, rfp
}

// insert inserts the item into filter
func insert(f *Filter, x []byte) (ok bool) {
	fp, fph := fingerprintOf(x, f.fingerprintSize, f.hash)
	i1, i2 := indicesOf(x, fph, f.totalBuckets, f.hash)

	defer func() {
		if ok {
			f.count++
		}
	}()

	if addToBucket(f.buckets[i1], fp) || addToBucket(f.buckets[i2], fp) {
		return true
	}

	is := []uint64{i1, i2}
	i1 = is[rand.Intn(len(is))]
	for k := 0; k < f.maxKicks; k++ {
		i1, fp = replaceItem(f, i1, fp)
		if addToBucket(f.buckets[i1], fp) {
			return true
		}
	}

	return false
}

// lookup checks if the item x existence in filter
func lookup(f *Filter, x []byte) bool {
	fp, fph := fingerprintOf(x, f.fingerprintSize, f.hash)
	i1, i2 := indicesOf(x, fph, f.totalBuckets, f.hash)

	if containsIn(f.buckets[i1], fp) || containsIn(f.buckets[i2], fp) {
		return true
	}

	return false
}

// deleteItem deletes item if present from the filter
func deleteItem(f *Filter, x []byte) (ok bool) {
	fp, fph := fingerprintOf(x, f.fingerprintSize, f.hash)
	i1, i2 := indicesOf(x, fph, f.totalBuckets, f.hash)

	defer func() {
		if ok {
			f.count--
		}
	}()

	if deleteFrom(f.buckets[i1], fp) || deleteFrom(f.buckets[i2], fp) {
		return true
	}

	return false
}

// Insert inserts the item to the filter
// returns error of filter is full
func (f *Filter) Insert(x []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return insert(f, x)
}

// InsertUnique inserts only unique items
func (f *Filter) InsertUnique(x []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return lookup(f, x) || insert(f, x)
}

// Lookup says if the given item exists in filter
func (f *Filter) Lookup(x []byte) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return lookup(f, x)
}

// Delete deletes the item from the filter
func (f *Filter) Delete(x []byte) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	return deleteItem(f, x)
}

// Count returns total inserted items into filter
func (f *Filter) Count() uint64 {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return f.count
}
