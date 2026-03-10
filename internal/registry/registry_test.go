package registry_test

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	"localgate/internal/registry"
)

func TestRegisterAndLookup(t *testing.T) {
	reg := registry.NewServiceRegistry()

	if err := reg.Register("foo", "localhost:3000", false); err != nil {
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

	if err := reg.Register("foo", "localhost:3000", false); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register("foo", "localhost:4000", true); err != nil {
		t.Fatal(err)
	}

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

	if err := reg.Register("foo", "localhost:3000", false); err != nil {
		t.Fatal(err)
	}
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

	err := reg.Register("", "localhost:3000", false)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRegisterEmptyTarget(t *testing.T) {
	reg := registry.NewServiceRegistry()

	err := reg.Register("foo", "", false)
	if err == nil {
		t.Fatal("expected error for empty target")
	}
}

func TestRegister_AlreadyExists(t *testing.T) {
	reg := registry.NewServiceRegistry()

	if err := reg.Register("foo", "localhost:3000", false); err != nil {
		t.Fatal(err)
	}
	err := reg.Register("foo", "localhost:4000", false)
	if err == nil {
		t.Fatal("expected ErrAlreadyExists, got nil")
	}
	if !errors.Is(err, registry.ErrAlreadyExists) {
		t.Errorf("expected ErrAlreadyExists, got: %v", err)
	}
	// 元のターゲットが変わっていないことを確認
	target, _ := reg.Lookup("foo")
	if target != "localhost:3000" {
		t.Errorf("expected original target 'localhost:3000', got '%s'", target)
	}
}

func TestRegister_ForceOverwrite(t *testing.T) {
	reg := registry.NewServiceRegistry()

	if err := reg.Register("foo", "localhost:3000", false); err != nil {
		t.Fatal(err)
	}
	err := reg.Register("foo", "localhost:4000", true)
	if err != nil {
		t.Fatalf("expected no error with force=true, got: %v", err)
	}
	target, _ := reg.Lookup("foo")
	if target != "localhost:4000" {
		t.Errorf("expected updated target 'localhost:4000', got '%s'", target)
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

	if err := reg.Register("foo", "localhost:3000", false); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register("bar", "localhost:4000", false); err != nil {
		t.Fatal(err)
	}

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
			_ = reg.Register(name, "localhost:3000", false)
		}(i)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("service%d", i)
			reg.Lookup(name)
		}(i)
	}
	wg.Wait()
}
