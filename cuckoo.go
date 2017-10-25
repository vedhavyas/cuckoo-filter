package cuckoo

import (
	"encoding/binary"
	"hash"
	"math/rand"
	"sync"

	"github.com/spaolacci/murmur3"
)

const (
	defaultBucketSize   = 4
	defaultTotalBuckets = 4 << 20
	defaultMaxKicks     = 500
	seed                = 59053
)

// fingerprint of the item
type fingerprint uint16

// emptyFingerprint represents an empty fingerprint
var emptyFingerprint fingerprint

// bucket with b fingerprints per bucket
type bucket []fingerprint

// tempBytes used to temporarily store the fingerprint
var tempBytes = make([]byte, 2, 2)

// Filter is the cuckoo-filter
type Filter struct {
	count        uint64
	buckets      []bucket
	bucketSize   uint8
	totalBuckets uint64
	hash         hash.Hash64
	maxKicks     uint16

	// protects above fields
	mu *sync.RWMutex
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
		buckets:      initBuckets(defaultTotalBuckets, defaultBucketSize),
		bucketSize:   defaultBucketSize,
		totalBuckets: defaultTotalBuckets,
		hash:         murmur3.New64WithSeed(seed),
		maxKicks:     defaultMaxKicks,
		mu:           &sync.RWMutex{},
	}
}

// deleteFrom deletes fingerprint from bucket if exists
func deleteFrom(b bucket, fp fingerprint) bool {
	for i := range b {
		if b[i] != fp {
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
		if b[i] == fp {
			return true
		}
	}

	return false
}

// addToBucket will add fp to the bucket i in filter
func addToBucket(b bucket, fp fingerprint) bool {
	for j := range b {
		if b[j] != emptyFingerprint {
			continue
		}

		b[j] = fp
		return true
	}

	return false
}

// hashOf returns the 64-bit hash
func hashOf(x []byte, hash hash.Hash64) uint64 {
	hash.Reset()
	hash.Write(x)
	return hash.Sum64()
}

// fingerprintOf returns the fingerprint of x with size using hash
func fingerprintOf(x []byte, hash hash.Hash64) (fp fingerprint, fph uint64) {
	ufp := binary.BigEndian.Uint16(x)
	return fingerprint(ufp), hashOf(x[:2], hash)
}

// indicesOf returns the indices of item x using given hash
func indicesOf(xh uint64, fph, totalBuckets uint64) (i1, i2 uint64) {
	i1 = xh % totalBuckets
	i2 = (i1 ^ fph) % totalBuckets
	return i1, i2
}

// replaceItem replaces fingerprint from i and returns the alternate index for kicked fingerprint
func replaceItem(f *Filter, i uint64, k int, fp fingerprint) (j uint64, rfp fingerprint) {
	rfp, f.buckets[i][k] = f.buckets[i][k], fp
	binary.BigEndian.PutUint16(tempBytes, uint16(fp))
	rfph := hashOf(tempBytes, f.hash)
	j = (i ^ rfph) % f.totalBuckets
	return j, rfp
}

// insert inserts the item into filter
func insert(f *Filter, x []byte) (ok bool) {
	fp, fph := fingerprintOf(x, f.hash)
	i1, i2 := indicesOf(hashOf(x, f.hash), fph, f.totalBuckets)

	defer func() {
		if ok {
			f.count++
		}
	}()

	if addToBucket(f.buckets[i1], fp) || addToBucket(f.buckets[i2], fp) {
		return true
	}

	rn := rand.Int()
	ri := []uint64{i1, i2}[rn%2]
	for k := uint16(0); k < f.maxKicks; k++ {
		ri, fp = replaceItem(f, ri, rn%int(f.bucketSize), fp)
		if addToBucket(f.buckets[i1], fp) {
			return true
		}
	}

	return false
}

// lookup checks if the item x existence in filter
func lookup(f *Filter, x []byte) bool {
	fp, fph := fingerprintOf(x, f.hash)
	i1, i2 := indicesOf(hashOf(x, f.hash), fph, f.totalBuckets)

	if containsIn(f.buckets[i1], fp) || containsIn(f.buckets[i2], fp) {
		return true
	}

	return false
}

// deleteItem deletes item if present from the filter
func deleteItem(f *Filter, x []byte) (ok bool) {
	fp, fph := fingerprintOf(x, f.hash)
	i1, i2 := indicesOf(hashOf(x, f.hash), fph, f.totalBuckets)

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

// check the bytes
func check(x []byte) ([]byte, bool) {
	if len(x) == 0 {
		return nil, false
	}

	if len(x) == 1 {
		x = []byte{0, x[0]}
	}

	return x, true
}

// Insert inserts the item to the filter
// returns error of filter is full
func (f *Filter) Insert(x []byte) bool {
	x, ok := check(x)
	if !ok {
		return false
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	return insert(f, x)
}

// InsertUnique inserts only unique items
func (f *Filter) InsertUnique(x []byte) bool {
	x, ok := check(x)
	if !ok {
		return false
	}

	if f.Lookup(x) {
		return true
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	return insert(f, x)
}

// Lookup says if the given item exists in filter
func (f *Filter) Lookup(x []byte) bool {
	x, ok := check(x)
	if !ok {
		return false
	}

	f.mu.RLock()
	defer f.mu.RUnlock()

	return lookup(f, x)
}

// Delete deletes the item from the filter
func (f *Filter) Delete(x []byte) bool {
	x, ok := check(x)
	if !ok {
		return false
	}

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
