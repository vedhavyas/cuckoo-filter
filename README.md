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

#### func (*Filter) Encode

```go
func (f *Filter) Encode(w io.Writer) error
```
Encode gob encodes the filter to passed writer

#### func Decode

```go
func Decode(r io.Reader) (*Filter, error)
```
Decode decodes and returns the filter instance


## Benchmarks

### 16 << 20 inserts with bucket size 4
```
Maximum Inserts
=================
Expected inserts: 16777216
Total inserted: 16022242
Load factor: 0.9550
Time Taken: 18.118615793s
=================


Lookups
=================
Total lookups: 16022242
Failed lookups: 0
Expected failed lookups: 0
Load factor: 0.9550
Time Taken: 11.394561881s
=================
```

### 16 << 20 inserts with bucket size 8
```
Maximum Inserts
=================
Expected inserts: 16777216
Total inserted: 16525558
Load factor: 0.9850
Time Taken: 17.478650065s
=================


Lookups
=================
Total lookups: 16525558
Failed lookups: 0
Expected failed lookups: 0
Load factor: 0.9850
Time Taken: 12.087359139s
=================

```

### 16 << 20 inserts with max bucket size(16)
```
Maximum Inserts
=================
Expected inserts: 16777216
Total inserted: 16676553
Load factor: 0.9940
Time Taken: 16.280342789s
=================


Lookups
=================
Total lookups: 16676553
Failed lookups: 0
Expected failed lookups: 0
Load factor: 0.9940
Time Taken: 12.45013918s
=================

```