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