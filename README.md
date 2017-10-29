# Cuckoo Filter

Practically better than Bloom Filter. Cuckoo filters support adding and removing items dynamically while achieving even higher performance than Bloom filters. For applications that store many items and target moderately low false positive rates, cuckoo filters have lower space overhead than space-optimized Bloom filters. It is a compact variant of a cuckoo hashtable that stores only fingerprints—a bit string derived from the item using a hash function—for each item inserted, instead of key-value pairs. The filter is densely filled with fingerprints (e.g., 95% entries occupied), which confers high space efficiency.

Paper can be found [here.](https://www.cs.cmu.edu/~dga/papers/cuckoo-conext2014.pdf)

--
    import "github.com/vedhavyas/cuckoo-filter"


## Usage

#### type Filter

```go
type Filter struct {
}
```

Filter is the cuckoo-filter

#### func  NewFilter

```go
func NewFilter(count uint32) *Filter
```

#### func  StdFilter

```go
func StdFilter() *Filter
```
StdFilter returns Standard Cuckoo-Filter

#### func (*Filter) Count

```go
func (f *Filter) Count() uint32
```
Count returns total inserted items into filter

#### func (*Filter) Delete

```go
func (f *Filter) Delete(x []byte) bool
```
Delete deletes the item from the filter

#### func (*Filter) Insert

```go
func (f *Filter) Insert(x []byte) bool
```
Insert inserts the item to the filter returns error of filter is full

#### func (*Filter) InsertUnique

```go
func (f *Filter) InsertUnique(x []byte) bool
```
InsertUnique inserts only unique items

#### func (*Filter) LoadFactor

```go
func (f *Filter) LoadFactor() float64
```
LoadFactor returns the load factor of the filter

#### func (*Filter) Lookup

```go
func (f *Filter) Lookup(x []byte) bool
```
Lookup says if the given item exists in filter


## Benchmarks

### 16 << 20 inserts with bucket size 4
```
  Maximum Inserts
 =================
 Expected inserts: 16777216
 Total inserted: 16025877
 Load factor: 0.9552
 First failure at: 16025877
 Time Taken: 31.103796543s
 =================
 
 
  Lookups
 =================
 Total lookups: 16025877
 Failed lookups: 1
 Expected failed lookups: 1 (Due to the final displacement)
 Load factor: 0.9552
 Time Taken: 19.477595601s
 =================

```

### 16 << 20 inserts with bucket size 8
```
 Maximum Inserts
=================
Expected inserts: 16777216
Total inserted: 16528881
Load factor: 0.9852
First failure at: 16528881
Time Taken: 27.556190215s
=================


 Lookups
=================
Total lookups: 16528881
Failed lookups: 1
Expected failed lookups: 0
Load factor: 0.9852
Time Taken: 26.037199544s
=================
```
