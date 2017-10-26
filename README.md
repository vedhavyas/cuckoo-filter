# Cuckoo Filter

Practically better than Bloom Filter. Paper can be found [here.](https://www.cs.cmu.edu/~dga/papers/cuckoo-conext2014.pdf)

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
Total inserted: 16047821
Load factor: 0.9565
First failure at: 16047821
Time Taken: 20.207010142s
=================


 Lookups
=================
Total lookups: 16047821
Failed lookups: 186
Expected failed lookups: 0
Load factor: 0.9565
Time Taken: 11.708892532s
=================

Res memory: 280M
```

### 16 << 20 inserts with bucket size 8
```
Maximum Inserts
=================
Expected inserts: 16777216
Total inserted: 16566681
Load factor: 0.9875
First failure at: 16566681
Time Taken: 18.237849646s
=================


 Lookups
=================
Total lookups: 16566681
Failed lookups: 215
Expected failed lookups: 0
Load factor: 0.9875
Time Taken: 11.798103504s
=================

Res Memory: 176M
```
