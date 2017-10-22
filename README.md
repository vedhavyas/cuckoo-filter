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

#### func  StdFilter

```go
func StdFilter() *Filter
```
StdFilter returns Standard Cuckoo-Filter

#### func (*Filter) Count

```go
func (f *Filter) Count() uint64
```
Count returns total inserted items into filter

#### func (*Filter) Delete

```go
func (f *Filter) Delete(x []byte) bool
```
Delete deletes the item from the filter

#### func (*Filter) Exists

```go
func (f *Filter) Exists(x []byte) bool
```
Exist says if the given items exists in filter

#### func (*Filter) Insert

```go
func (f *Filter) Insert(x []byte) error
```
Insert inserts the item to the filter returns error of filter is full
