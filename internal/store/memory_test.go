package store

import (
	"testing"

	"github.com/mrckurz/CI-CD-MCM/internal/model"
)

func TestCreateAndGet(t *testing.T) {
	s := NewMemoryStore()
	created := s.Create(model.Product{Name: "Widget", Price: 9.99})

	got, err := s.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID(%d) returned unexpected error: %v", created.ID, err)
	}
	if got.ID != created.ID {
		t.Errorf("ID = %d, want %d", got.ID, created.ID)
	}
	if got.Name != "Widget" {
		t.Errorf("Name = %q, want %q", got.Name, "Widget")
	}
	if got.Price != 9.99 {
		t.Errorf("Price = %v, want %v", got.Price, 9.99)
	}
}

func TestUpdateProduct(t *testing.T) {
	s := NewMemoryStore()
	created := s.Create(model.Product{Name: "Old", Price: 1.00})

	updated, err := s.Update(created.ID, model.Product{Name: "New", Price: 2.00})
	if err != nil {
		t.Fatalf("Update(%d) returned unexpected error: %v", created.ID, err)
	}
	if updated.Name != "New" {
		t.Errorf("Name = %q, want %q", updated.Name, "New")
	}
	if updated.Price != 2.00 {
		t.Errorf("Price = %v, want %v", updated.Price, 2.00)
	}

	got, err := s.GetByID(created.ID)
	if err != nil {
		t.Fatalf("GetByID(%d) after update returned unexpected error: %v", created.ID, err)
	}
	if got.Name != "New" || got.Price != 2.00 {
		t.Errorf("GetByID after update = {%q, %v}, want {%q, %v}", got.Name, got.Price, "New", 2.00)
	}
}

func TestDeleteProduct(t *testing.T) {
	s := NewMemoryStore()
	created := s.Create(model.Product{Name: "ToDelete", Price: 5.00})

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("Delete(%d) returned unexpected error: %v", created.ID, err)
	}

	_, err := s.GetByID(created.ID)
	if err != ErrNotFound {
		t.Errorf("GetByID(%d) after delete: got err = %v, want ErrNotFound", created.ID, err)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	tests := []struct {
		name string
		id   int
	}{
		{"negative ID", -1},
		{"zero ID", 0},
		{"non-existent ID", 999},
	}

	s := NewMemoryStore()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := s.GetByID(tc.id)
			if err != ErrNotFound {
				t.Errorf("GetByID(%d) err = %v, want ErrNotFound", tc.id, err)
			}
		})
	}
}

func TestGetAllEmpty(t *testing.T) {
	s := NewMemoryStore()
	products := s.GetAll()
	if len(products) != 0 {
		t.Errorf("expected 0 products, got %d", len(products))
	}
}

func TestDeleteNonExistent(t *testing.T) {
	s := NewMemoryStore()
	err := s.Delete(999)
	if err != ErrNotFound {
		t.Error("expected ErrNotFound when deleting non-existent product")
	}
}
