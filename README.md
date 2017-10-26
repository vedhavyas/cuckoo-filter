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
