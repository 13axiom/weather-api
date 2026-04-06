package service_test

import (
	"testing"

	"github.com/13axiom/weather-api/internal/service"
)

func TestComputeHash_SameDataProducesSameHash(t *testing.T) {
	h1 := service.ComputeHash(20.5, 5.0, 0.0, 1, "2024-01-01T12:00")
	h2 := service.ComputeHash(20.5, 5.0, 0.0, 1, "2024-01-01T12:00")
	if h1 != h2 {
		t.Errorf("same data must produce same hash: got %q and %q", h1, h2)
	}
}

func TestComputeHash_DifferentDataProducesDifferentHash(t *testing.T) {
	h1 := service.ComputeHash(20.5, 5.0, 0.0, 1, "2024-01-01T12:00")
	h2 := service.ComputeHash(21.0, 5.0, 0.0, 1, "2024-01-01T12:00") // temp changed
	if h1 == h2 {
		t.Error("different temperature must produce different hash")
	}
}

func TestComputeHash_DifferentTimeProducesDifferentHash(t *testing.T) {
	h1 := service.ComputeHash(20.5, 5.0, 0.0, 1, "2024-01-01T12:00")
	h2 := service.ComputeHash(20.5, 5.0, 0.0, 1, "2024-01-01T13:00") // time changed
	if h1 == h2 {
		t.Error("different time must produce different hash")
	}
}

func TestComputeHash_NotEmpty(t *testing.T) {
	h := service.ComputeHash(0, 0, 0, 0, "2024-01-01T00:00")
	if h == "" {
		t.Error("hash must not be empty")
	}
}
