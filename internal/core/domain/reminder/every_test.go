package reminder

import (
	"testing"
	"time"
)

func TestEveryPredefinedVars(t *testing.T) {
	if !EveryDay.IsValid() {
		t.Fatal()
	}
	if !EveryWeek.IsValid() {
		t.Fatal()
	}
	if !EveryMonth.IsValid() {
		t.Fatal()
	}
	if !EveryMonth.IsValid() {
		t.Fatal()
	}
}

func TestEveryValid(t *testing.T) {
	_, ok := NewEvery(time.Minute)
	if !ok {
		t.Fatal()
	}
	_, ok = NewEvery(time.Hour * 366)
	if !ok {
		t.Fatal()
	}
	_, ok = NewEvery(time.Hour * 72)
	if !ok {
		t.Fatal()
	}
	_, ok = NewEvery(time.Minute * 5)
	if !ok {
		t.Fatal()
	}
}
