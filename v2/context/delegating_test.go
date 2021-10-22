package context

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestValuesDelegating(t *testing.T) {
	type key string
	tests := []struct {
		name   string
		child  context.Context
		parent context.Context
		assert func(*testing.T, context.Context)
	}{
		{
			name:   "it delegates to child first",
			child:  context.WithValue(context.Background(), key("foo"), "foo"),
			parent: context.WithValue(context.Background(), key("foo"), "bar"),
			assert: func(t *testing.T, c context.Context) {
				if v := c.Value(key("foo")); v != "foo" {
					t.Errorf("expected child value, got %s", v)
				}
			},
		},
		{
			name:   "it delegates to parent if missing from child",
			child:  context.Background(),
			parent: context.WithValue(context.Background(), key("foo"), "foo"),
			assert: func(t *testing.T, c context.Context) {
				if v := c.Value(key("foo")); v != "foo" {
					t.Errorf("expected parent value, got %s", v)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValuesDelegating(tt.child, tt.parent)
			tt.assert(t, got)
		})
	}
}
func TestValuesDelegatingDelegatesOtherwiseToChild(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())
	child, childCancel := context.WithCancel(context.Background())
	derived := ValuesDelegating(child, parent)

	ch := make(chan string)
	go func() {
		<-derived.Done()
		ch <- "derived"
	}()
	go func() {
		<-child.Done()
		ch <- "child"
	}()
	go func() {
		<-parent.Done()
		ch <- "parent"
	}()

	parentCancel()
	v1 := <-ch
	if v1 != "parent" {
		t.Errorf("cancelling parent should not cancel child or derived: %s", v1)
	}
	childCancel()
	v2 := <-ch
	v3 := <-ch
	diff := cmp.Diff([]string{"derived", "child"}, []string{v2, v3}, cmpopts.SortSlices(func(a, b string) bool { return a < b }))
	if diff != "" {
		t.Errorf("unexpected (-want, +got) = %v", diff)
	}
}
