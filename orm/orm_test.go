package main

import "testing"

func BenchmarkGorp(b *testing.B) {
	initPq()
	b.Log("N=", b.N)
	for n := 0; n < b.N; n++ {
		testgorp()
	}
}

func BenchmarkPq(b *testing.B) {
	initPq()
	b.Log("N=", b.N)
	for n := 0; n < b.N; n++ {
		testpq()
	}
}

func BenchmarkPgx(b *testing.B) {
	initPgx()
	b.Log("N=", b.N)
	for n := 0; n < b.N; n++ {
		testpgx()
	}
}

func BenchmarkGoPg(b *testing.B) {
	initGoPg()
	b.Log("N=", b.N)
	for n := 0; n < b.N; n++ {
		testGoPg()
	}
}
