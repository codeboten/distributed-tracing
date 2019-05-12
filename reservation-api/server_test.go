package main

import "testing"

func TestAbs(t *testing.T) {
	s := Spaceship{tankSize: 10}
	got := s.maxFuelQuantity()
	if got != 10 {
		t.Errorf("maxFuelQuantity() = %d; want 10", got)
	}
	refuel(&s)
}
