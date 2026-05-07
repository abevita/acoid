use acoid::{
    from_digest, generate, generate_str, is_supported_length, is_valid, validate, validate_length,
    Error,
};

// ── Story 3: validate_length / is_supported_length ────────────────────────────

#[test]
fn test_is_supported_length_valid() {
    for l in [6, 8, 10, 12] {
        assert!(is_supported_length(l), "expected {l} to be supported");
    }
}

#[test]
fn test_is_supported_length_invalid() {
    for l in [0, 1, 5, 7, 9, 11, 13, 16] {
        assert!(!is_supported_length(l), "expected {l} to be unsupported");
    }
}

#[test]
fn test_validate_length_valid() {
    for l in [6, 8, 10, 12] {
        assert!(validate_length(l).is_ok(), "length {l} should be valid");
    }
}

#[test]
fn test_validate_length_invalid() {
    for l in [0, 5, 7, 16] {
        match validate_length(l) {
            Err(Error::UnsupportedLength(n)) => assert_eq!(n, l),
            other => panic!("expected UnsupportedLength for {l}, got {other:?}"),
        }
    }
}

// ── Story 1: generate ─────────────────────────────────────────────────────────

#[test]
fn test_generate_deterministic() {
    let input = b"test-input";
    for l in [6, 8, 10, 12] {
        let a = generate(input, l).unwrap();
        let b = generate(input, l).unwrap();
        assert_eq!(a, b, "generate should be deterministic at length {l}");
    }
}

#[test]
fn test_generate_output_length() {
    let input = b"output-length-test";
    for l in [6, 8, 10, 12] {
        let v = generate(input, l).unwrap();
        assert_eq!(v.len(), l, "output length should be {l}");
    }
}

#[test]
fn test_generate_id57_charset() {
    let banned = "0oOIl";
    let input = b"charset-check";
    for l in [6, 8, 10, 12] {
        let v = generate(input, l).unwrap();
        for ch in banned.chars() {
            assert!(
                !v.contains(ch),
                "output {v:?} at length {l} contains banned char '{ch}'"
            );
        }
    }
}

#[test]
fn test_generate_different_inputs_differ() {
    let a = generate(b"alpha", 8).unwrap();
    let b = generate(b"beta", 8).unwrap();
    assert_ne!(a, b, "different inputs should produce different ACOIDs");
}

#[test]
fn test_generate_unsupported_length() {
    assert!(matches!(
        generate(b"x", 7),
        Err(Error::UnsupportedLength(7))
    ));
}

// ── Story 2: from_digest ──────────────────────────────────────────────────────

#[test]
fn test_from_digest_deterministic() {
    let digest: Vec<u8> = (0u8..32).collect();
    for l in [6, 8, 10, 12] {
        let a = from_digest(&digest, l).unwrap();
        let b = from_digest(&digest, l).unwrap();
        assert_eq!(a, b, "from_digest should be deterministic at length {l}");
    }
}

#[test]
fn test_from_digest_output_length() {
    let digest: Vec<u8> = (0u8..32).map(|i| i.wrapping_mul(3)).collect();
    for l in [6, 8, 10, 12] {
        let v = from_digest(&digest, l).unwrap();
        assert_eq!(v.len(), l);
    }
}

#[test]
fn test_from_digest_no_banned_chars() {
    let banned = "0oOIl";
    let digest: Vec<u8> = (0u8..32).map(|i| i.wrapping_mul(7)).collect();
    for l in [6, 8, 10, 12] {
        let v = from_digest(&digest, l).unwrap();
        for ch in banned.chars() {
            assert!(!v.contains(ch), "from_digest output {v:?} contains banned char '{ch}'");
        }
    }
}

#[test]
fn test_from_digest_unsupported_length() {
    assert!(matches!(
        from_digest(&[0u8; 32], 9),
        Err(Error::UnsupportedLength(9))
    ));
}

// ── Story 4: validate / is_valid ─────────────────────────────────────────────

#[test]
fn test_validate_valid_acoids() {
    for l in [6, 8, 10, 12] {
        let v = generate(b"validate-roundtrip", l).unwrap();
        assert!(validate(&v, l).is_ok(), "validate({v:?}, {l}) should pass");
    }
}

#[test]
fn test_validate_banned_chars() {
    let cases = [("A0BCDE", '0'), ("AoBCDE", 'o'), ("AOBCDE", 'O'), ("AIBCDE", 'I'), ("AlBCDE", 'l')];
    for (value, ch) in cases {
        assert!(
            matches!(validate(value, 6), Err(Error::InvalidCharset)),
            "validate({value:?}) should reject banned char '{ch}'"
        );
    }
}

#[test]
fn test_validate_length_mismatch() {
    assert!(matches!(
        validate("ABCDE", 6),
        Err(Error::LengthMismatch { got: 5, want: 6 })
    ));
}

#[test]
fn test_validate_unsupported_length() {
    assert!(matches!(
        validate("ABCDEFG", 7),
        Err(Error::UnsupportedLength(7))
    ));
}

#[test]
fn test_is_valid_true() {
    let v = generate(b"isvalid-true", 8).unwrap();
    assert!(is_valid(&v, 8));
}

#[test]
fn test_is_valid_false_banned_char() {
    assert!(!is_valid("AB0CDE", 6));
}

// ── Story 5: generate_str / parity ───────────────────────────────────────────

#[test]
fn test_generate_str_parity() {
    let input = "string-bytes-parity";
    for l in [6, 8, 10, 12] {
        let from_bytes = generate(input.as_bytes(), l).unwrap();
        let from_str = generate_str(input, l).unwrap();
        assert_eq!(from_bytes, from_str, "parity failure at length {l}");
    }
}

#[test]
fn test_generate_empty_input_deterministic() {
    let a = generate(b"", 8).unwrap();
    let b = generate(b"", 8).unwrap();
    assert_eq!(a, b);
    assert_eq!(a.len(), 8);
}

// ── Cross-language stable vectors ─────────────────────────────────────────────
//
// These vectors were established from the Go reference implementation.
// All language implementations MUST produce these exact outputs.

#[test]
fn test_stable_vectors() {
    let cases = [
        ("acoid:stable:v1", 6, "MyZx2x"),
        ("acoid:stable:v1", 8, "MyZx2x9Y"),
        ("acoid:stable:v1", 10, "MyZx2x9YGy"),
        ("acoid:stable:v1", 12, "MyZx2x9YGyMr"),
    ];
    for (input, length, want) in cases {
        let got = generate_str(input, length).unwrap();
        assert_eq!(
            got, want,
            "stable vector broken: input={input:?} length={length} want={want:?} got={got:?}"
        );
    }
}

// ── E2E round-trip ────────────────────────────────────────────────────────────

#[test]
fn test_e2e_full_round_trip() {
    let cases = [
        ("user:abc123", 6),
        ("venue:central-park", 6),
        ("sport:pickleball", 6),
        ("player:jane-doe", 8),
        ("match:2026-05-01:court-3", 10),
        ("activity:drill:forehand", 10),
        ("pair:match:abc:A", 10),
        ("result:match:abc:final", 10),
        ("actpart:activity:xyz:player:p1", 12),
        ("submission:match:abc:round:1", 12),
    ];
    for (input, length) in cases {
        let id = generate_str(input, length).unwrap();
        assert_eq!(id.len(), length, "length mismatch for {input:?}");
        assert!(validate(&id, length).is_ok(), "validate failed for {input:?}");
        // deterministic
        let id2 = generate_str(input, length).unwrap();
        assert_eq!(id, id2, "not deterministic for {input:?}");
    }
}
