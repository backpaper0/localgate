package registry_test

import (
	"fmt"
	"sync"
	"testing"

	"localgate/internal/registry"
)

func TestRegisterAndLookup(t *testing.T) {
	reg := registry.NewServiceRegistry()

	if err := reg.Register("foo", "localhost:3000"); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	target, found := reg.Lookup("foo")
	if !found {
		t.Fatal("expected to find registered service")
	}
	if target != "localhost:3000" {
		t.Errorf("expected target 'localhost:3000', got '%s'", target)
	}
}

func TestRegisterOverwrite(t *testing.T) {
	reg := registry.NewServiceRegistry()

	reg.Register("foo", "localhost:3000")
	reg.Register("foo", "localhost:4000")

	target, found := reg.Lookup("foo")
	if !found {
		t.Fatal("expected to find registered service")
	}
	if target != "localhost:4000" {
		t.Errorf("expected updated target 'localhost:4000', got '%s'", target)
	}
}

func TestDeregister(t *testing.T) {
	reg := registry.NewServiceRegistry()

	reg.Register("foo", "localhost:3000")
	if err := reg.Deregister("foo"); err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}

	_, found := reg.Lookup("foo")
	if found {
		t.Fatal("expected service to be deregistered")
	}
}

func TestDeregisterNotFound(t *testing.T) {
	reg := registry.NewServiceRegistry()

	err := reg.Deregister("nonexistent")
	if err == nil {
		t.Fatal("expected error for deregistering nonexistent service")
	}
}

func TestRegisterEmptyName(t *testing.T) {
	reg := registry.NewServiceRegistry()

	err := reg.Register("", "localhost:3000")
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegisterEmptyTarget(t *testing.T) {
	reg := registry.NewServiceRegistry()

	err := reg.Register("foo", "")
	if err == nil {
		t.Fatal("expected error for empty target")
	}
}

func TestListEmpty(t *testing.T) {
	reg := registry.NewServiceRegistry()

	entries := reg.List()
	if len(entries) != 0 {
		t.Errorf("expected empty list, got %d entries", len(entries))
	}
}

func TestList(t *testing.T) {
	reg := registry.NewServiceRegistry()

	reg.Register("foo", "localhost:3000")
	reg.Register("bar", "localhost:4000")

	entries := reg.List()
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestLookupNotFound(t *testing.T) {
	reg := registry.NewServiceRegistry()

	_, found := reg.Lookup("unknown")
	if found {
		t.Fatal("expected not found for unregistered service")
	}
}

func TestConcurrentAccess(t *testing.T) {
	reg := registry.NewServiceRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("service%d", i)
			reg.Register(name, "localhost:3000")
		}(i)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("service%d", i)
			reg.Lookup(name)
		}(i)
	}
	wg.Wait()
}
