package client

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudevents/sdk-go/pkg/cloudevents"
	"github.com/prometheus/common/log"
)

const watchLag = 10 * time.Millisecond

func TestNewDefaultOverrides(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir, err := ioutil.TempDir("", "*")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	defaulter, err := NewDefaultOverrides(ctx, dir)
	if err != nil {
		t.Error(err)
	}
	if defaulter == nil {
		t.Fatal("expected defaulter function to not be nil")
	}

	key1 := "foo"
	value1 := "bar"
	file1 := filepath.Join(dir, key1)

	if err := ioutil.WriteFile(file1, []byte(value1), 0666); err != nil {
		log.Fatal(err)
	}
	time.Sleep(watchLag) // Gives time for the watcher to notice.

	got := defaulter(ctx, cloudevents.New("1.0"))

	if got.Extensions()[key1] != value1 {
		t.Fatalf("expected %s to be %s, but got %s", key1, value1, got.Extensions()[key1])
	}
}

func TestNewOverridesObserver(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	dir, err := ioutil.TempDir("", "overrides")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	key1 := "foo"
	value1 := "scary"
	file1 := filepath.Join(dir, key1)

	if err := ioutil.WriteFile(file1, []byte(value1), 0666); err != nil {
		log.Fatal(err)
	}

	o := NewOverridesObserver(dir)

	// Start tests.

	// Test 1: files that already existed before the watcher.
	{
		got := o.Apply(ctx, cloudevents.New("1.0"))

		if got.Extensions()[key1] != value1 {
			t.Fatalf("expected %s to be %s, but got %s", key1, value1, got.Extensions()[key1])
		}
	}
	t.Log("ok - test 1")

	// Start the watcher.
	go func() {
		if err := o.Watch(ctx); err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(watchLag) // Gives time for the watcher to notice.

	// Update the watched directory and confirm the new key also is added.

	key2 := "bar"
	value2 := "bottles"

	if err := ioutil.WriteFile(filepath.Join(dir, key2), []byte(value2), 0666); err != nil {
		log.Fatal(err)
	}
	time.Sleep(watchLag) // Gives time for the watcher to notice.

	{ // Test 2: files that are created after the watcher.
		got := o.Apply(ctx, cloudevents.New("1.0"))

		if got.Extensions()[key1] != value1 {
			t.Fatalf("expected %s to be %s, but got %s", key1, value1, got.Extensions()[key1])
		}
		if got.Extensions()[key2] != value2 {
			t.Fatalf("expected %s to be %s, but got %s", key2, value2, got.Extensions()[key2])
		}
	}
	t.Log("ok - test 2")

	// Change one of the files.
	value2Delta := "take one down"

	if err := ioutil.WriteFile(filepath.Join(dir, key2), []byte(value2Delta), 0666); err != nil {
		log.Fatal(err)
	}
	time.Sleep(watchLag) // Gives time for the watcher to notice.

	{ // Test 3: files mutated while being watched.
		got := o.Apply(ctx, cloudevents.New("1.0"))

		if got.Extensions()[key1] != value1 {
			t.Fatalf("expected %s to be %s, but got %s", key1, value1, got.Extensions()[key1])
		}
		if got.Extensions()[key2] != value2Delta {
			t.Fatalf("expected %s to be %s, but got %s", key2, value2Delta, got.Extensions()[key2])
		}
	}
	t.Log("ok - test 3")

	// Now delete key1
	if err := os.Remove(file1); err != nil {
		t.Fatalf("failed to delete file %s, %s", file1, err)
	}
	time.Sleep(2 * watchLag) // Gives time for the watcher to notice.

	{ // Test 4: delete a file while being watched.
		got := o.Apply(ctx, cloudevents.New("1.0"))

		if _, ok := got.Extensions()[key1]; ok {
			t.Fatalf("expected %s to be nil, but got %s", key1, got.Extensions()[key1])
		}
		if got.Extensions()[key2] != value2Delta {
			t.Fatalf("expected %s to be %s, but got %s", key2, value2Delta, got.Extensions()[key2])
		}
	}
	t.Log("ok - test 4")

	// Create a file with a bad key name.

	key3 := "foo-bar"
	value3 := "this is not a valid key"

	if err := ioutil.WriteFile(filepath.Join(dir, key3), []byte(value3), 0666); err != nil {
		log.Fatal(err)
	}
	time.Sleep(watchLag) // Gives time for the watcher to notice.

	{ // Test 5: files that are created after the watcher.
		got := o.Apply(ctx, cloudevents.New("1.0"))

		if _, ok := got.Extensions()[key1]; ok {
			t.Fatalf("expected %s to be nil, but got %s", key1, got.Extensions()[key1])
		}
		if got.Extensions()[key2] != value2Delta {
			t.Fatalf("expected %s to be %s, but got %s", key2, value2Delta, got.Extensions()[key2])
		}
		if _, ok := got.Extensions()[key3]; ok {
			t.Fatalf("expected %s to be nil, but got %s", key3, got.Extensions()[key3])
		}
	}
	t.Log("ok - test 5")

	// Test 6: the string method.

	got := o.String()
	want := fmt.Sprintf("%s\n", `bar: "take one down"`)
	if want != got {
		t.Fatalf(".String(), expected %s, got %s", want, got)
	}
	t.Log("ok - test 6")
}
