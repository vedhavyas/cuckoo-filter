package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/vedhavyas/cuckoo-filter"
)

func loadMax(f *cuckoo.Filter, n uint32) (tt string, ff uint32) {
	start := time.Now()
	for i := uint32(0); i < n; i++ {
		x := []byte(fmt.Sprintf("test-%d", i))
		if !f.Insert(x) {
			ff = f.Count()
			break
		}

	}

	return time.Now().Sub(start).String(), ff
}

func lookup(f *cuckoo.Filter, n uint32) (tt string, mw uint32) {
	start := time.Now()
	for i := uint32(0); i < n; i++ {
		x := []byte(fmt.Sprintf("test-%d", i))
		if !f.Lookup(x) {
			mw++
		}
	}

	return time.Now().Sub(start).String(), mw
}

func main() {
	var n uint32 = 16 << 20
	f, err := cuckoo.NewFilterWithBucketSize(n, 16)
	if err != nil {
		log.Fatal(err)
	}
	tt, _ := loadMax(f, n)
	fmt.Println(" Maximum Inserts")
	fmt.Println("=================")
	fmt.Printf("Expected inserts: %d\n", n)
	fmt.Printf("Total inserted: %d\n", f.Count())
	fmt.Printf("Load factor: %0.4f\n", f.LoadFactor())
	fmt.Printf("Time Taken: %s\n", tt)
	fmt.Println("=================")
	fmt.Print("\n\n")

	tt, mw := lookup(f, f.Count())
	fmt.Println(" Lookups")
	fmt.Println("=================")
	fmt.Printf("Total lookups: %d\n", f.Count())
	fmt.Printf("Failed lookups: %d\n", mw)
	fmt.Printf("Expected failed lookups: %d\n", 0)
	fmt.Printf("Load factor: %0.4f\n", f.LoadFactor())
	fmt.Printf("Time Taken: %s\n", tt)
	fmt.Println("=================")

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt)
	for range ch {
		return
	}
}
