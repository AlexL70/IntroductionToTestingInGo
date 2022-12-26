package main

import (
	"testing"
)

func Test_isPrime(t *testing.T) {
	primeTests := []struct {
		name     string
		num      int
		expected bool
		msg      string
	}{
		{"prime", 7, true, "7 is a prime number!"},
		{"not prime", 8, false, "8 is not prime because it is divisible by 2!"},
	}
	for _, e := range primeTests {
		result, msg := isPrime(e.num)
		if e.expected != result {
			t.Errorf("%s: expected %t but got %t", e.name, e.expected, result)
		}
		if e.msg != msg {
			t.Errorf("%s: expected \"%s\" but got \"%s\"", e.name, e.msg, msg)
		}
	}
}
