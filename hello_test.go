package main

import (
	"testing"
)

func TestSayHello(t *testing.T) {
	got := SayHello("Julia")
	want := "Hello, Julia!"

	if got != want {
		t.Errorf("got: %s, want: %s", got, want)
	}
}
