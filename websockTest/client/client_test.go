package client

import (
	"testing"
)

func Benchmark_Client(b *testing.B) {
	for i := 0; i < b.N; i++ { //use b.N for looping
		client1()
	}
}
