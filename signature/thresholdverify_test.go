package signature

import (
	"testing"
)

func TestVerify(test *testing.T) {
	n := 3

	message := "Test"

	suite, pubPoly, sigShares, err := GenerateKeyShares(n, []byte(message))

	if err != nil {
		test.Fatal(err)
	}
	err = VerifyMessage(n, []byte(message), suite, pubPoly, sigShares)
	if err != nil {
		test.Fatal(err)
	}
}
