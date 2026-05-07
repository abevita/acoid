package acoid_test

import (
	"fmt"

	"github.com/abevita/acoid/go"
)

func ExampleGenerateString() {
	id, err := acoid.GenerateString("user:123", 6)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(id) == 6)
	// Output:
	// true
}

func ExampleGenerate() {
	id, err := acoid.Generate([]byte("match:abc:1"), 10)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(id) == 10)
	// Output:
	// true
}

func ExampleValidate() {
	id, _ := acoid.GenerateString("venue:xyz", 6)
	err := acoid.Validate(id, 6)
	fmt.Println(err == nil)
	// Output:
	// true
}

func ExampleValidateLength() {
	fmt.Println(acoid.ValidateLength(8) == nil)
	fmt.Println(acoid.ValidateLength(7) == nil)
	// Output:
	// true
	// false
}

func ExampleIsSupportedLength() {
	fmt.Println(acoid.IsSupportedLength(10))
	fmt.Println(acoid.IsSupportedLength(11))
	// Output:
	// true
	// false
}

func ExampleFromDigest() {
	digest := make([]byte, 32)
	id, err := acoid.FromDigest(digest, 8)
	if err != nil {
		panic(err)
	}
	fmt.Println(len(id) == 8)
	fmt.Println(acoid.IsValid(id, 8))
	// Output:
	// true
	// true
}

func ExampleIsValid() {
	id, _ := acoid.GenerateString("player:jane", 8)
	fmt.Println(acoid.IsValid(id, 8))
	fmt.Println(acoid.IsValid("AB0CDE12", 8)) // '0' is banned
	// Output:
	// true
	// false
}
