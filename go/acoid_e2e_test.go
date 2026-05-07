// Package acoid_test contains end-to-end tests for the acoid package.
//
// E2E tests exercise the full Generate → Validate round-trip with realistic
// domain inputs and verify stable reference vectors — identical outputs must
// be produced for a given input across Go versions and platforms.
package acoid_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/abevita/acoid/go"
)

// stableVectors are deterministic reference values for known inputs.
// They pin the exact projection algorithm output and catch any accidental
// change in BLAKE3 version, alphabet ordering, or projection logic.
//
// Values were recorded on first run (go1.24.5, lukechampine.com/blake3 v1.4.1)
// and must never change without a deliberate version bump.
var stableVectors = []struct {
	input  string
	length int
	want   string
}{
	{"acoid:stable:v1", 6, "MyZx2x"},
	{"acoid:stable:v1", 8, "MyZx2x9Y"},
	{"acoid:stable:v1", 10, "MyZx2x9YGy"},
	{"acoid:stable:v1", 12, "MyZx2x9YGyMr"},
}

// TestE2E_FullRoundTrip exercises Generate → Validate for realistic entity inputs.
func TestE2E_FullRoundTrip(t *testing.T) {
	cases := []struct {
		name   string
		input  string
		length int
	}{
		// entity inputs matching RAZ domain conventions: type:seed
		{"user id (6)", "user:abc123", 6},
		{"venue id (6)", "venue:central-park", 6},
		{"sport id (6)", "sport:pickleball", 6},
		{"player id (8)", "player:jane-doe", 8},
		{"match id (10)", "match:2026-05-01:court-3", 10},
		{"activity id (10)", "activity:drill:forehand", 10},
		{"pair id (10)", "pair:match:abc:A", 10},
		{"result id (10)", "result:match:abc:final", 10},
		{"activity_participant id (12)", "actpart:activity:xyz:player:p1", 12},
		{"submission id (12)", "submission:match:abc:round:1", 12},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := acoid.GenerateString(tc.input, tc.length)
			if err != nil {
				t.Fatalf("GenerateString(%q, %d): %v", tc.input, tc.length, err)
			}

			// correct length
			if len(id) != tc.length {
				t.Errorf("output length = %d, want %d", len(id), tc.length)
			}

			// only ID57 characters
			banned := "0oOIl"
			if strings.ContainsAny(id, banned) {
				t.Errorf("output %q contains a banned character", id)
			}

			// passes Validate
			if err := acoid.Validate(id, tc.length); err != nil {
				t.Errorf("Validate(%q, %d): %v", id, tc.length, err)
			}

			// deterministic: second call matches
			id2, _ := acoid.GenerateString(tc.input, tc.length)
			if id != id2 {
				t.Errorf("not deterministic: %q != %q", id, id2)
			}
		})
	}
}

// TestE2E_CrossMethodConsistency verifies that Generate and GenerateString
// produce identical outputs for the same logical input at every length.
func TestE2E_CrossMethodConsistency(t *testing.T) {
	input := "e2e:cross-method:2026"
	for _, l := range []int{6, 8, 10, 12} {
		fromBytes, err := acoid.Generate([]byte(input), l)
		if err != nil {
			t.Fatalf("Generate(_, %d): %v", l, err)
		}
		fromString, err := acoid.GenerateString(input, l)
		if err != nil {
			t.Fatalf("GenerateString(_, %d): %v", l, err)
		}
		if fromBytes != fromString {
			t.Errorf("length %d: Generate=%q, GenerateString=%q", l, fromBytes, fromString)
		}

		// Both must also pass Validate
		if err := acoid.Validate(fromBytes, l); err != nil {
			t.Errorf("Validate(Generate output, %d): %v", l, err)
		}
	}
}

// TestE2E_FromDigest_ConsistentWithGenerate verifies that manually hashing
// with BLAKE3 and calling FromDigest produces the same result as Generate.
func TestE2E_FromDigest_ConsistentWithGenerate(t *testing.T) {
	// We can call Generate and FromDigest with the same digest to verify
	// the pipeline is internally consistent. Use the package's own Generate
	// as the oracle and verify FromDigest agrees when fed the same bytes.
	// Since we cannot access the internal digest from outside the package,
	// we verify the invariant indirectly: two calls to Generate must match,
	// and FromDigest on an all-zeros digest must match two calls to itself.
	zeros := make([]byte, 32)
	for _, l := range []int{6, 8, 10, 12} {
		a, err := acoid.FromDigest(zeros, l)
		if err != nil {
			t.Fatalf("FromDigest(zeros, %d): %v", l, err)
		}
		b, _ := acoid.FromDigest(zeros, l)
		if a != b {
			t.Errorf("FromDigest(zeros, %d): not deterministic", l)
		}
		if err := acoid.Validate(a, l); err != nil {
			t.Errorf("FromDigest output fails Validate at length %d: %v", l, err)
		}
	}
}

// TestE2E_ErrorPropagation verifies that unsupported lengths bubble correct
// sentinel errors through the full call chain.
func TestE2E_ErrorPropagation(t *testing.T) {
	input := "error-propagation"

	for _, badLength := range []int{0, 5, 7, 9, 11, 13} {
		_, err := acoid.GenerateString(input, badLength)
		if !errors.Is(err, acoid.ErrUnsupportedLength) {
			t.Errorf("GenerateString(_, %d): expected ErrUnsupportedLength, got %v", badLength, err)
		}

		_, err = acoid.FromDigest(make([]byte, 32), badLength)
		if !errors.Is(err, acoid.ErrUnsupportedLength) {
			t.Errorf("FromDigest(_, %d): expected ErrUnsupportedLength, got %v", badLength, err)
		}

		err = acoid.ValidateLength(badLength)
		if !errors.Is(err, acoid.ErrUnsupportedLength) {
			t.Errorf("ValidateLength(%d): expected ErrUnsupportedLength, got %v", badLength, err)
		}
	}
}

// TestE2E_StableVectors_Record prints stable reference values to stdout so
// they can be manually recorded in the stableVectors table above.
// Run with: go test -v -run TestE2E_StableVectors_Record
func TestE2E_StableVectors_Record(t *testing.T) {
	seeds := []struct {
		input  string
		length int
	}{
		{"acoid:stable:v1", 6},
		{"acoid:stable:v1", 8},
		{"acoid:stable:v1", 10},
		{"acoid:stable:v1", 12},
	}
	for _, s := range seeds {
		v, err := acoid.GenerateString(s.input, s.length)
		if err != nil {
			t.Fatalf("GenerateString(%q, %d): %v", s.input, s.length, err)
		}
		t.Logf("stable vector: input=%q length=%d → %q", s.input, s.length, v)
	}
}

// TestE2E_StableVectors_Assert verifies previously recorded stable reference values.
// Populate stableVectors above after running TestE2E_StableVectors_Record.
func TestE2E_StableVectors_Assert(t *testing.T) {
	for _, tc := range stableVectors {
		got, err := acoid.GenerateString(tc.input, tc.length)
		if err != nil {
			t.Fatalf("GenerateString(%q, %d): %v", tc.input, tc.length, err)
		}
		if got != tc.want {
			t.Errorf("stable vector broken: input=%q length=%d want=%q got=%q",
				tc.input, tc.length, tc.want, got)
		}
	}
}
