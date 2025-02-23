package gomap_test

import (
	"context"
	"testing"

	"go.autokitteh.dev/autokitteh/internal/kvstore"
	"go.autokitteh.dev/autokitteh/internal/kvstore/encoding"
	"go.autokitteh.dev/autokitteh/internal/kvstore/gomap"
	"go.autokitteh.dev/autokitteh/internal/kvstore/test"
)

// TestStore tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
func TestStore(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store := createStore(t, encoding.JSON)
		test.TestStore(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store := createStore(t, encoding.Gob)
		test.TestStore(store, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store := createStore(t, encoding.JSON)
		test.TestTypes(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store := createStore(t, encoding.Gob)
		test.TestTypes(store, t)
	})
}

// TestStoreConcurrent launches a bunch of goroutines that concurrently work with one store.
// The store is Go map with manual locking via sync.RWMutex, so testing this is important.
func TestStoreConcurrent(t *testing.T) {
	store := createStore(t, encoding.JSON)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, store)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// Test empty key
	store := createStore(t, encoding.JSON)
	err := store.Set(context.TODO(), "", "bar")
	if err == nil {
		t.Error("Expected an error")
	}
	_, err = store.Get(context.TODO(), "", new(string))
	if err == nil {
		t.Error("Expected an error")
	}
	err = store.Delete(context.TODO(), "")
	if err == nil {
		t.Error("Expected an error")
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
func TestNil(t *testing.T) {
	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		store := createStore(t, encoding.JSON)
		err := store.Set(context.TODO(), "foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		store := createStore(t, encoding.Gob)
		err := store.Set(context.TODO(), "foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			store := createStore(t, codec)

			// Prep
			err := store.Set(context.TODO(), "foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = store.Get(context.TODO(), "foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i any // actually nil
			_, err = store.Get(context.TODO(), "foo", i)
			if err == nil {
				t.Error("An error was expected")
			}

			var valPtr *test.Foo // nil value
			_, err = store.Get(context.TODO(), "foo", valPtr)
			if err == nil {
				t.Error("An error was expected")
			}
		}
	}
	t.Run("get with nil / nil value parameter", createTest(encoding.JSON))
	t.Run("get with nil / nil value parameter", createTest(encoding.Gob))
}

// TestClose tests if the close method returns any errors.
func TestClose(t *testing.T) {
	store := createStore(t, encoding.JSON)
	err := store.Close()
	if err != nil {
		t.Error(err)
	}
}

func createStore(t *testing.T, codec encoding.Codec) kvstore.Store {
	options := gomap.Options{
		Codec: codec,
	}
	store := gomap.NewStore(options)
	return store
}
