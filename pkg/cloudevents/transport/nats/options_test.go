package nats

import (
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"testing"
)

func TestWithEncoding(t *testing.T) {
	testCases := map[string]struct {
		t        *Transport
		encoding Encoding
		want     *Transport
		wantErr  string
	}{
		"valid encoding": {
			t:        &Transport{},
			encoding: StructuredV03,
			want: &Transport{
				Encoding: StructuredV03,
			},
		},
	}
	for n, tc := range testCases {
		t.Run(n, func(t *testing.T) {

			err := tc.t.applyOptions(WithEncoding(tc.encoding))

			if tc.wantErr != "" || err != nil {
				var gotErr string
				if err != nil {
					gotErr = err.Error()
				}
				if diff := cmp.Diff(tc.wantErr, gotErr); diff != "" {
					t.Errorf("unexpected error (-want, +got) = %v", diff)
				}
				return
			}

			got := tc.t

			if diff := cmp.Diff(tc.want, got,
				cmpopts.IgnoreUnexported(Transport{})); diff != "" {
				t.Errorf("unexpected (-want, +got) = %v", diff)
			}
		})
	}
}
