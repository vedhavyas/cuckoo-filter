package cuckoo

import (
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"hash"
	"io"
	"math/rand"
	"sync"

	"github.com/spaolacci/murmur3"
)

const (
	defaultBucketSize   = 8
	maxBucketSize       = 16
	defaultTotalBuckets = 4 << 20
	defaultMaxKicks     = 500
	seed                = 59053
)

// fingerprint of the item
type fingerprint uint16

// emptyFingerprint
var emptyFingerprint fingerprint

// bucket with n fingerprints
type bucket struct {
	Track uint16
	FPs   []fingerprint
}

// Filter is the cuckoo-filter
type Filter struct {
	count        uint32
	buckets      []bucket
	bucketSize   uint8
	totalBuckets uint32
	hash         hash.Hash32
	maxKicks     uint16

	// protects above fields
	L sync.RWMutex
}

// gobFilter for encoding and decoding the Filter
type gobFilter struct {
	Count        uint32
	Buckets      []bucket
	BucketSize   uint8
	TotalBuckets uint32
	MaxKicks     uint16
}

// initBuckets initialises the buckets
func initBuckets(totalBuckets uint32, bucketSize uint8) []bucket {
	buckets := make([]bucket, totalBuckets, totalBuckets)
	for i := range buckets {
		buckets[i] = bucket{FPs: make([]fingerprint, bucketSize, bucketSize)}
	}

	return buckets
}

// StdFilter returns Standard Cuckoo-Filter
func StdFilter() *Filter {
	return newFilter(defaultTotalBuckets, defaultBucketSize, murmur3.New32WithSeed(seed))
}

func newFilter(tb uint32, bs uint8, hash hash.Hash32) *Filter {
	return &Filter{
		buckets:      initBuckets(tb, bs),
		bucketSize:   bs,
		totalBuckets: tb,
		hash:         hash,
		maxKicks:     defaultMaxKicks,
	}
}

func NewFilter(count uint32) *Filter {
	b := nextPowerOf2(count) / defaultBucketSize
	return newFilter(b, defaultBucketSize, murmur3.New32WithSeed(seed))
}

func NewFilterWithBucketSize(count uint32, bs uint8) (*Filter, error) {
	if bs > maxBucketSize {
		return nil, fmt.Errorf("doesn't support %d bucket size. Max bucket size is %d", bs, maxBucketSize)
	}

	b := nextPowerOf2(count) / uint32(bs)
	return newFilter(b, bs, murmur3.New32WithSeed(seed)), nil
}

// nextPowerOf2 returns the next power 2 >= v
func nextPowerOf2(v uint32) (n uint32) {
	var i uint32
	for i = 2; i < 32; i++ {
		n = 1 << i
		if n >= v {
			break
		}
	}

	return n
}

// isSet returns true if the i th bit in the Track is 1
func isSet(track uint16, i uint8) bool {
	return track|(1<<i) == track
}

// unSet un sets the i th bit in Track
func unSet(track uint16, i uint8) uint16 {
	return track ^ 1<<i
}

func set(track uint16, i uint8) uint16 {
	return track | 1<<i
}

// deleteFrom deletes fingerprint from bucket if exists
func deleteFrom(b *bucket, bs uint8, fp fingerprint) bool {
	for i := uint8(0); i < bs; i++ {
		if !isSet(b.Track, i) || b.FPs[i] != fp {
			continue
		}

		b.FPs[i] = emptyFingerprint
		b.Track = unSet(b.Track, i)
		return true
	}

	return false
}

// containsIn returns if the given fingerprint exists in bucket
func containsIn(b bucket, bs uint8, fp fingerprint) bool {
	for i := uint8(0); i < bs; i++ {
		if isSet(b.Track, i) && b.FPs[i] == fp {
			return true
		}
	}

	return false
}

// addToBucket will add fp to the bucket i in filter
func addToBucket(b *bucket, bs uint8, fp fingerprint) bool {
	for i := uint8(0); i < bs; i++ {
		if isSet(b.Track, i) {
			continue
		}

		b.FPs[i] = fp
		b.Track = set(b.Track, i)
		return true
	}

	return false
}

// hashOf returns the 32-bit hash
func hashOf(x []byte, hash hash.Hash32) (uint32, []byte) {
	hash.Reset()
	hash.Write(x)
	h := hash.Sum32()
	return h, hash.Sum(nil)
}

// fingerprintOf returns the fingerprint of x with size using hash
func fingerprintOf(xb []byte) (fp fingerprint) {
	return fingerprint(binary.BigEndian.Uint16(xb))
}

func fingerprintHash(fp fingerprint, hash hash.Hash32) (fph uint32) {
	b := make([]byte, 2, 2)
	binary.BigEndian.PutUint16(b, uint16(fp))
	fph, _ = hashOf(b, hash)
	return fph
}

// indicesOf returns the indices of item x using given hash
func indicesOf(xh, fph, totalBuckets uint32) (i1, i2 uint32) {
	i1 = xh % totalBuckets
	i2 = alternateIndex(totalBuckets, i1, fph)
	return i1, i2
}

// alternateIndex returns the alternate index of i
func alternateIndex(totalBuckets, i, fph uint32) (j uint32) {
	return (i ^ fph) % totalBuckets
}

// estimatedLoadFactor returns an estimated max load factor based on bucket size
func estimatedLoadFactor(bucketSize uint8) float64 {
	switch {
	case bucketSize < 8:
		return 0.955
	case bucketSize < 16:
		return 0.985
	default:
		return 0.994
	}
}

// isReliable returns if the filter is reliable for another insert
func isReliable(f *Filter) bool {
	clf := f.ULoadFactor()
	elf := estimatedLoadFactor(f.bucketSize)
	if clf < elf {
		return true
	}

	return false
}

// swapFingerprint swaps a random fp from the bucket with provided fp
func swapFingerprint(b *bucket, fp fingerprint) fingerprint {
	var sfp fingerprint
	k := rand.Intn(len(b.FPs))
	sfp, b.FPs[k] = b.FPs[k], fp
	return sfp
}

// insert inserts the item into filter
func insert(f *Filter, x []byte) (ok bool) {
	xh, xb := hashOf(x, f.hash)
	fp := fingerprintOf(xb)
	fph := fingerprintHash(fp, f.hash)
	i1, i2 := indicesOf(xh, fph, f.totalBuckets)

	defer func() {
		if ok {
			f.count++
		}
	}()

	if addToBucket(&f.buckets[i1], f.bucketSize, fp) || addToBucket(&f.buckets[i2], f.bucketSize, fp) {
		return true
	}

	ri := []uint32{i1, i2}[rand.Intn(2)]
	var k uint16
	for k = 0; k < f.maxKicks; k++ {
		fp = swapFingerprint(&f.buckets[ri], fp)
		fph = fingerprintHash(fp, f.hash)
		ri = alternateIndex(f.totalBuckets, ri, fph)
		if addToBucket(&f.buckets[ri], f.bucketSize, fp) {
			return true
		}
	}

	return false
}

// lookup checks if the item x existence in filter
func lookup(f *Filter, x []byte) bool {
	xh, xb := hashOf(x, f.hash)
	fp := fingerprintOf(xb)
	fph := fingerprintHash(fp, f.hash)
	i1, i2 := indicesOf(xh, fph, f.totalBuckets)

	if containsIn(f.buckets[i1], f.bucketSize, fp) || containsIn(f.buckets[i2], f.bucketSize, fp) {
		return true
	}

	return false
}

// deleteItem deletes item if present from the filter
func deleteItem(f *Filter, x []byte) (ok bool) {
	xh, xb := hashOf(x, f.hash)
	fp := fingerprintOf(xb)
	fph := fingerprintHash(fp, f.hash)
	i1, i2 := indicesOf(xh, fph, f.totalBuckets)

	defer func() {
		if ok {
			f.count--
		}
	}()

	if deleteFrom(&f.buckets[i1], f.bucketSize, fp) || deleteFrom(&f.buckets[i2], f.bucketSize, fp) {
		return true
	}

	return false
}

// sanitize the bytes
func sanitize(x []byte) ([]byte, bool) {
	if len(x) == 0 {
		return nil, false
	}

	if len(x) == 1 {
		x = []byte{0, x[0]}
	}

	return x, true
}

// Insert inserts the item to the filter
func (f *Filter) Insert(x []byte) bool {
	f.L.Lock()
	defer f.L.Unlock()

	return f.UInsert(x)
}

// UInsert inserts the item to the filter. Not thread safe
func (f *Filter) UInsert(x []byte) bool {
	if !isReliable(f) {
		return false
	}

	x, ok := sanitize(x)
	if !ok {
		return false
	}

	return insert(f, x)
}

// InsertUnique inserts only unique items
func (f *Filter) InsertUnique(x []byte) bool {
	f.L.Lock()
	defer f.L.Unlock()

	return f.UInsertUnique(x)
}

// UInsertUnique inserts only unique items. Not thread safe
func (f *Filter) UInsertUnique(x []byte) bool {
	if !isReliable(f) {
		return false
	}

	x, ok := sanitize(x)
	if !ok {
		return false
	}

	if f.ULookup(x) {
		return true
	}

	return insert(f, x)
}

// Lookup checks if item exists in filter
func (f *Filter) Lookup(x []byte) bool {
	f.L.RLock()
	defer f.L.RUnlock()

	return f.ULookup(x)
}

// ULookup checks if item exists in filter. Not thread safe
func (f *Filter) ULookup(x []byte) bool {
	x, ok := sanitize(x)
	if !ok {
		return false
	}

	return lookup(f, x)
}

// Delete deletes the item from the filter
func (f *Filter) Delete(x []byte) bool {
	f.L.Lock()
	defer f.L.Unlock()

	return f.UDelete(x)
}

// UDelete deletes the item from the filter. Not thread safe
func (f *Filter) UDelete(x []byte) bool {
	x, ok := sanitize(x)
	if !ok {
		return false
	}

	return deleteItem(f, x)
}

// Count returns total inserted items into filter
func (f *Filter) Count() uint32 {
	f.L.RLock()
	defer f.L.RUnlock()

	return f.UCount()
}

// UCount returns total inserted items into filter. Not thread safe
func (f *Filter) UCount() uint32 {
	return f.count
}

// LoadFactor returns the load factor of the filter
func (f *Filter) LoadFactor() float64 {
	f.L.RLock()
	defer f.L.RUnlock()
	return f.ULoadFactor()
}

// ULoadFactor returns the load factor of the filter. Not thread safe
func (f *Filter) ULoadFactor() float64 {
	return float64(f.count) / (float64(uint32(f.bucketSize) * f.totalBuckets))
}

// Encode gob encodes the filter to passed writer
func (f *Filter) Encode(w io.Writer) error {
	// hold the read lock till we encode the data to the writer
	f.L.RLock()
	defer f.L.RUnlock()
	gf := &gobFilter{
		Count:        f.count,
		Buckets:      f.buckets,
		BucketSize:   f.bucketSize,
		TotalBuckets: f.totalBuckets,
		MaxKicks:     f.maxKicks,
	}

	ge := gob.NewEncoder(w)
	return ge.Encode(gf)
}

// Decode decodes and returns the filter instance
func Decode(r io.Reader) (*Filter, error) {
	gd := gob.NewDecoder(r)
	gf := &gobFilter{}
	err := gd.Decode(gf)
	if err != nil {
		return nil, fmt.Errorf("failed to decode filter: %v", err)
	}

	f := &Filter{
		count:        gf.Count,
		buckets:      gf.Buckets,
		bucketSize:   gf.BucketSize,
		totalBuckets: gf.TotalBuckets,
		hash:         murmur3.New32WithSeed(seed),
		maxKicks:     gf.MaxKicks,
	}

	return f, nil
}
