package acoid_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/abevita/acoid"
)

// ── Story 3: ValidateLength / IsSupportedLength ───────────────────────────────

func TestIsSupportedLength_Supported(t *testing.T) {
	for _, l := range []int{6, 8, 10, 12} {
		if !acoid.IsSupportedLength(l) {
			t.Errorf("IsSupportedLength(%d) = false, want true", l)
		}
	}
}

func TestIsSupportedLength_Unsupported(t *testing.T) {
	for _, l := range []int{0, 1, 5, 7, 9, 11, 16, -1} {
		if acoid.IsSupportedLength(l) {
			t.Errorf("IsSupportedLength(%d) = true, want false", l)
		}
	}
}

func TestValidateLength_Supported(t *testing.T) {
	for _, l := range []int{6, 8, 10, 12} {
		if err := acoid.ValidateLength(l); err != nil {
			t.Errorf("ValidateLength(%d): unexpected error: %v", l, err)
		}
	}
}

func TestValidateLength_Unsupported(t *testing.T) {
	cases := []int{0, 5, 7, 16}
	for _, l := range cases {
		err := acoid.ValidateLength(l)
		if err == nil {
			t.Errorf("ValidateLength(%d): expected error, got nil", l)
			continue
		}
		if !errors.Is(err, acoid.ErrUnsupportedLength) {
			t.Errorf("ValidateLength(%d): error type = %T, want ErrUnsupportedLength", l, err)
		}
	}
}

// ── Story 1: Generate ─────────────────────────────────────────────────────────

func TestGenerate_Deterministic(t *testing.T) {
	input := []byte("test-input")
	for _, l := range []int{6, 8, 10, 12} {
		a, err := acoid.Generate(input, l)
		if err != nil {
			t.Fatalf("Generate(_, %d): %v", l, err)
		}
		b, _ := acoid.Generate(input, l)
		if a != b {
			t.Errorf("Generate(_, %d): not deterministic: %q != %q", l, a, b)
		}
	}
}

func TestGenerate_OutputLength(t *testing.T) {
	input := []byte("output-length-test")
	for _, l := range []int{6, 8, 10, 12} {
		v, err := acoid.Generate(input, l)
		if err != nil {
			t.Fatalf("Generate(_, %d): %v", l, err)
		}
		if len(v) != l {
			t.Errorf("Generate(_, %d): output length = %d, want %d", l, len(v), l)
		}
	}
}

func TestGenerate_ID57Charset(t *testing.T) {
	banned := "0oOIl"
	input := []byte("charset-check")
	for _, l := range []int{6, 8, 10, 12} {
		v, _ := acoid.Generate(input, l)
		if strings.ContainsAny(v, banned) {
			t.Errorf("Generate(_, %d): output %q contains a banned character", l, v)
		}
	}
}

func TestGenerate_DifferentInputs_DifferentOutputs(t *testing.T) {
	a, _ := acoid.Generate([]byte("alpha"), 8)
	b, _ := acoid.Generate([]byte("beta"), 8)
	if a == b {
		t.Errorf("Generate: different inputs produced the same ACOID: %q", a)
	}
}

func TestGenerate_UnsupportedLength(t *testing.T) {
	_, err := acoid.Generate([]byte("x"), 7)
	if err == nil {
		t.Fatal("Generate with unsupported length: expected error, got nil")
	}
	if !errors.Is(err, acoid.ErrUnsupportedLength) {
		t.Errorf("Generate: error type = %T, want ErrUnsupportedLength", err)
	}
}

// ── Story 1: MustGenerate ────────────────────────────────────────────────────

func TestMustGenerate_Valid(t *testing.T) {
	v := acoid.MustGenerate([]byte("must-generate"), 8)
	if len(v) != 8 {
		t.Errorf("MustGenerate: length = %d, want 8", len(v))
	}
}

func TestMustGenerate_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGenerate with unsupported length: expected panic, got none")
		}
	}()
	acoid.MustGenerate([]byte("x"), 7)
}

// ── Story 2: FromDigest ───────────────────────────────────────────────────────

func TestFromDigest_Deterministic(t *testing.T) {
	digest := make([]byte, 32)
	for i := range digest {
		digest[i] = byte(i)
	}
	for _, l := range []int{6, 8, 10, 12} {
		a, err := acoid.FromDigest(digest, l)
		if err != nil {
			t.Fatalf("FromDigest(_, %d): %v", l, err)
		}
		b, _ := acoid.FromDigest(digest, l)
		if a != b {
			t.Errorf("FromDigest(_, %d): not deterministic", l)
		}
	}
}

func TestFromDigest_OutputLength(t *testing.T) {
	digest := make([]byte, 32)
	for i := range digest {
		digest[i] = byte(i * 3)
	}
	for _, l := range []int{6, 8, 10, 12} {
		v, err := acoid.FromDigest(digest, l)
		if err != nil {
			t.Fatalf("FromDigest(_, %d): %v", l, err)
		}
		if len(v) != l {
			t.Errorf("FromDigest(_, %d): output length = %d, want %d", l, len(v), l)
		}
	}
}

func TestFromDigest_NoBannedChars(t *testing.T) {
	banned := "0oOIl"
	digest := make([]byte, 32)
	for i := range digest {
		digest[i] = byte(i * 7)
	}
	for _, l := range []int{6, 8, 10, 12} {
		v, _ := acoid.FromDigest(digest, l)
		if strings.ContainsAny(v, banned) {
			t.Errorf("FromDigest(_, %d): output %q contains a banned character", l, v)
		}
	}
}

func TestFromDigest_UnsupportedLength(t *testing.T) {
	_, err := acoid.FromDigest(make([]byte, 32), 9)
	if !errors.Is(err, acoid.ErrUnsupportedLength) {
		t.Errorf("FromDigest: expected ErrUnsupportedLength, got %v", err)
	}
}

// ── Story 4: Validate / IsValid ───────────────────────────────────────────────

func TestValidate_ValidACOIDs(t *testing.T) {
	for _, l := range []int{6, 8, 10, 12} {
		v, _ := acoid.Generate([]byte("validate-roundtrip"), l)
		if err := acoid.Validate(v, l); err != nil {
			t.Errorf("Validate(%q, %d): unexpected error: %v", v, l, err)
		}
	}
}

func TestValidate_BannedChars(t *testing.T) {
	cases := []struct {
		name  string
		value string
	}{
		{"digit zero", "A0BCDE"},
		{"lower o", "AoBCDE"},
		{"upper O", "AOBCDE"},
		{"upper I", "AIBCDE"},
		{"lower l", "AlBCDE"},
	}
	for _, tc := range cases {
		err := acoid.Validate(tc.value, 6)
		if err == nil {
			t.Errorf("Validate (%s): expected error, got nil", tc.name)
			continue
		}
		if !errors.Is(err, acoid.ErrInvalidCharset) {
			t.Errorf("Validate (%s): error type = %T, want ErrInvalidCharset", tc.name, err)
		}
	}
}

func TestValidate_LengthMismatch(t *testing.T) {
	err := acoid.Validate("ABCDE", 6) // 5 chars, want 6
	if err == nil {
		t.Fatal("Validate with wrong char count: expected error, got nil")
	}
	if !errors.Is(err, acoid.ErrLengthMismatch) {
		t.Errorf("Validate: error type = %T, want ErrLengthMismatch", err)
	}
}

func TestValidate_UnsupportedLength(t *testing.T) {
	err := acoid.Validate("ABCDEFG", 7)
	if err == nil {
		t.Fatal("Validate with unsupported length 7: expected error, got nil")
	}
	if !errors.Is(err, acoid.ErrUnsupportedLength) {
		t.Errorf("Validate: error type = %T, want ErrUnsupportedLength", err)
	}
}

func TestIsValid_True(t *testing.T) {
	v, _ := acoid.Generate([]byte("isvalid-true"), 8)
	if !acoid.IsValid(v, 8) {
		t.Errorf("IsValid(%q, 8) = false, want true", v)
	}
}

func TestIsValid_False_BannedChar(t *testing.T) {
	if acoid.IsValid("AB0CDE", 6) {
		t.Error("IsValid with banned char '0': expected false, got true")
	}
}

func TestIsValid_False_UnsupportedLength(t *testing.T) {
	if acoid.IsValid("ABCDEFG", 7) {
		t.Error("IsValid with unsupported length 7: expected false, got true")
	}
}

// ── Story 5: String / bytes API parity ───────────────────────────────────────

func TestGenerateString_ParityWithGenerate(t *testing.T) {
	input := "string-bytes-parity"
	for _, l := range []int{6, 8, 10, 12} {
		fromBytes, _ := acoid.Generate([]byte(input), l)
		fromString, err := acoid.GenerateString(input, l)
		if err != nil {
			t.Fatalf("GenerateString(_, %d): %v", l, err)
		}
		if fromBytes != fromString {
			t.Errorf("length %d: Generate vs GenerateString differ: %q vs %q", l, fromBytes, fromString)
		}
	}
}

func TestGenerate_EmptyInput_Deterministic(t *testing.T) {
	a, err := acoid.Generate([]byte{}, 8)
	if err != nil {
		t.Fatalf("Generate(empty, 8): %v", err)
	}
	if len(a) != 8 {
		t.Errorf("Generate(empty, 8): output length = %d, want 8", len(a))
	}
	b, _ := acoid.Generate([]byte{}, 8)
	if a != b {
		t.Error("Generate(empty): not deterministic")
	}
}

func TestGenerateString_EmptyMatchesByte(t *testing.T) {
	fromByte, _ := acoid.Generate([]byte(""), 8)
	fromStr, _ := acoid.GenerateString("", 8)
	if fromByte != fromStr {
		t.Errorf("empty byte vs empty string: %q != %q", fromByte, fromStr)
	}
}

func TestGenerateString_OutputIsID57(t *testing.T) {
	banned := "0oOIl"
	for _, l := range []int{6, 8, 10, 12} {
		v, _ := acoid.GenerateString("id57-check", l)
		if strings.ContainsAny(v, banned) {
			t.Errorf("GenerateString(_, %d): output %q contains a banned character", l, v)
		}
	}
}
