package entity

import "testing"

const uuid = "15f79ac3-5049-418e-87dc-d4622ec40c30"

var secret = []byte("secret")

func BenchmarkBuildJWTString(b *testing.B) {
	for range b.N {
		//nolint:errcheck
		BuildJWTString(uuid, secret)
	}
}

func BenchmarkParseJWTString(b *testing.B) {
	jwt, err := BuildJWTString(uuid, secret)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()

	for range b.N {
		//nolint:errcheck
		ParseJWTString(jwt, secret)
	}
}
