package main

import (
	"fmt"
	"time"

	"github.com/vedhavyas/cuckoo-filter"
)

func loadMax(f *cuckoo.Filter, n uint32) (tt string, mw uint32, ff uint32, items [][]byte) {
	start := time.Now()
	var c bool
	for i := uint32(0); i < n; i++ {
		x := []byte(fmt.Sprintf("test-%d", i))
		if !f.Insert(x) {
			mw++
			if !c {
				ff = f.Count()
				c = true
			}
		}

		items = append(items, x)
	}

	return time.Now().Sub(start).String(), mw, ff, items
}

func lookup(f *cuckoo.Filter, items [][]byte) (tt string, mw uint32) {
	start := time.Now()
	for _, x := range items {
		if !f.Lookup(x) {
			mw++
		}
	}

	return time.Now().Sub(start).String(), mw
}

func main() {
	var n uint32 = 16 << 20
	f := cuckoo.NewFilter(n)
	tt, mw, ff, items := loadMax(f, n)
	fmt.Println(" Maximum Inserts")
	fmt.Println("=================")
	fmt.Printf("Total inserted: %d\n", f.Count())
	fmt.Printf("Load factor: %0.4f\n", f.LoadFactor())
	fmt.Printf("Failed inserts: %d\n", mw)
	fmt.Printf("First failure at: %d\n", ff)
	fmt.Printf("Time Taken: %s\n", tt)
	fmt.Println("=================")
	fmt.Print("\n\n")

	tt, mw = lookup(f, items)
	fmt.Println(" Lookups")
	fmt.Println("=================")
	fmt.Printf("Total lookups: %d\n", n)
	fmt.Printf("Load factor: %0.4f\n", f.LoadFactor())
	fmt.Printf("Failed lookups: %d\n", mw)
	fmt.Printf("Time Taken: %s\n", tt)
	fmt.Println("=================")
}
