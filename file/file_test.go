package file_test

import (
	"log"
	"os"
	"testing"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/encoding"
	"github.com/philippgille/gokv/file"
	"github.com/philippgille/gokv/test"
)

// TestStore tests if reading from, writing to and deleting from the store works properly.
// A struct is used as value. See TestTypes() for a test that is simpler but tests all types.
func TestStore(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store, path := createStore(t, encoding.JSON)
		defer cleanUp(store, path)
		test.TestStore(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store, path := createStore(t, encoding.Gob)
		defer cleanUp(store, path)
		test.TestStore(store, t)
	})
}

// TestTypes tests if setting and getting values works with all Go types.
func TestTypes(t *testing.T) {
	// Test with JSON
	t.Run("JSON", func(t *testing.T) {
		store, path := createStore(t, encoding.JSON)
		defer cleanUp(store, path)
		test.TestTypes(store, t)
	})

	// Test with gob
	t.Run("gob", func(t *testing.T) {
		store, path := createStore(t, encoding.Gob)
		defer cleanUp(store, path)
		test.TestTypes(store, t)
	})
}

// TestStoreConcurrent launches a bunch of goroutines that concurrently work with one store.
// The store is Go map with manual locking via sync.RWMutex, so testing this is important.
func TestStoreConcurrent(t *testing.T) {
	store, path := createStore(t, encoding.JSON)
	defer cleanUp(store, path)

	goroutineCount := 1000

	test.TestConcurrentInteractions(t, goroutineCount, store)
}

// TestErrors tests some error cases.
func TestErrors(t *testing.T) {
	// Test empty key
	store, path := createStore(t, encoding.JSON)
	defer cleanUp(store, path)
	err := store.Set("", "bar")
	if err == nil {
		t.Error("Expected an error")
	}
	_, err = store.Get("", new(string))
	if err == nil {
		t.Error("Expected an error")
	}
	err = store.Delete("")
	if err == nil {
		t.Error("Expected an error")
	}
}

// TestNil tests the behaviour when passing nil or pointers to nil values to some methods.
func TestNil(t *testing.T) {
	// Test setting nil

	t.Run("set nil with JSON marshalling", func(t *testing.T) {
		store, path := createStore(t, encoding.JSON)
		defer cleanUp(store, path)
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	t.Run("set nil with Gob marshalling", func(t *testing.T) {
		store, path := createStore(t, encoding.Gob)
		defer cleanUp(store, path)
		err := store.Set("foo", nil)
		if err == nil {
			t.Error("Expected an error")
		}
	})

	// Test passing nil or pointer to nil value for retrieval

	createTest := func(codec encoding.Codec) func(t *testing.T) {
		return func(t *testing.T) {
			store, path := createStore(t, codec)
			defer cleanUp(store, path)

			// Prep
			err := store.Set("foo", test.Foo{Bar: "baz"})
			if err != nil {
				t.Error(err)
			}

			_, err = store.Get("foo", nil) // actually nil
			if err == nil {
				t.Error("An error was expected")
			}

			var i any // actually nil
			_, err = store.Get("foo", i)
			if err == nil {
				t.Error("An error was expected")
			}

			var valPtr *test.Foo // nil value
			_, err = store.Get("foo", valPtr)
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
	store, path := createStore(t, encoding.JSON)
	defer func() { _ = os.RemoveAll(path) }()
	err := store.Close()
	if err != nil {
		t.Error(err)
	}
}

func createStore(t *testing.T, codec encoding.Codec) (file.Store, string) {
	path := generateRandomTempDBpath(t)
	options := file.Options{
		Directory: path,
		// Setting no FileNameEnding leads to all Codecs writing ".json" files,
		// but that doesn't matter to the functionality of gokv.
		Codec: codec,
	}
	store, err := file.NewStore(options)
	if err != nil {
		t.Fatal(err)
	}
	return store, path
}

func generateRandomTempDBpath(t *testing.T) string {
	path, err := os.MkdirTemp(os.TempDir(), "gokv")
	if err != nil {
		t.Fatalf("Generating random DB path failed: %v", err)
	}
	return path
}

// cleanUp cleans up the store (deletes the files that have been created during a test).
// If an error occurs the test is NOT marked as failed.
func cleanUp(store gokv.Store, path string) {
	err := store.Close()
	if err != nil {
		log.Printf("Error during cleaning up after a test (during closing the store): %v\n", err)
	}
	err = os.RemoveAll(path)
	if err != nil {
		log.Printf("Error during cleaning up after a test (during removing the data directory): %v\n", err)
	}
}
