package store

import (
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

func TestCreateAndGet(t *testing.T) {
	s := NewMemoryStore()
	p := s.Create(model.Product{Name: "Widget", Price: 9.99})

	got, err := s.GetByID(p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Widget" || got.Price != 9.99 {
		t.Errorf("got %+v, want Name=Widget Price=9.99", got)
	}
}

func TestGetAllEmpty(t *testing.T) {
	s := NewMemoryStore()
	products := s.GetAll()
	if len(products) != 0 {
		t.Errorf("expected 0 products, got %d", len(products))
	}
}

func TestGetAllWithProducts(t *testing.T) {
	s := NewMemoryStore()
	s.Create(model.Product{Name: "A", Price: 1.0})
	s.Create(model.Product{Name: "B", Price: 2.0})

	products := s.GetAll()
	if len(products) != 2 {
		t.Errorf("expected 2 products, got %d", len(products))
	}
}

func TestGetByIDNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.GetByID(999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateAssignsID(t *testing.T) {
	s := NewMemoryStore()
	p1 := s.Create(model.Product{Name: "A", Price: 1.0})
	p2 := s.Create(model.Product{Name: "B", Price: 2.0})

	if p1.ID >= p2.ID {
		t.Errorf("expected sequential IDs, got %d and %d", p1.ID, p2.ID)
	}
}

func TestUpdateExisting(t *testing.T) {
	s := NewMemoryStore()
	p := s.Create(model.Product{Name: "Old", Price: 1.0})

	updated, err := s.Update(p.ID, model.Product{Name: "New", Price: 2.0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Name != "New" || updated.Price != 2.0 {
		t.Errorf("got %+v, want Name=New Price=2.0", updated)
	}
	if updated.ID != p.ID {
		t.Errorf("expected ID %d, got %d", p.ID, updated.ID)
	}
}

func TestUpdateNotFound(t *testing.T) {
	s := NewMemoryStore()
	_, err := s.Update(999, model.Product{Name: "Ghost", Price: 1.0})
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeleteExisting(t *testing.T) {
	s := NewMemoryStore()
	p := s.Create(model.Product{Name: "Widget", Price: 9.99})

	err := s.Delete(p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = s.GetByID(p.ID)
	if err != ErrNotFound {
		t.Error("expected product to be deleted")
	}
}

func TestDeleteNonExistent(t *testing.T) {
	s := NewMemoryStore()
	err := s.Delete(999)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound when deleting non-existent product")
	}
}
