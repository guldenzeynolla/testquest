package main

import (
	"io/ioutil"
	"testing"
)

func BenchmarkNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		New(ioutil.Discard)
	}
}

func BenchmarkOld(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Old(ioutil.Discard)
	}
}

//go test -bench . -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1
