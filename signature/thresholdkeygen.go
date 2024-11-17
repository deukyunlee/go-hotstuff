package signature

import (
	"deukyunlee/hotstuff/logging"
	"go.dedis.ch/kyber/v4/pairing/bn256"
	"go.dedis.ch/kyber/v4/share"
	"go.dedis.ch/kyber/v4/sign/tbls"
	"math"
)

var (
	logger = logging.GetLogger()
)

func GenerateKeyShares(n int, msg []byte) (*bn256.Suite, *share.PubPoly, [][]byte, error) {
	// Generally, t is n * 2/3 in multi-sig system
	t := int(math.Ceil(float64(n) / 3 * 2))

	// Initialize a cryptographic suite using bn256, which supports pairing-based cryptography
	suite := bn256.NewSuite()

	secret := suite.G1().Scalar().Pick(suite.RandomStream())

	priPoly := share.NewPriPoly(suite.G2(), t, secret, suite.RandomStream())

	pubPoly := priPoly.Commit(suite.G2().Point().Base())

	sigShares := make([][]byte, 0)

	for i, x := range priPoly.Shares(n) {

		sig, err := tbls.Sign(suite, x, msg)
		if err != nil {
			return nil, nil, nil, err
		}
		logger.Infof("Share %d: %x\n", x, sig)

		sigShares = append(sigShares, sig)

		if i >= t {
			logger.Info("Threshold reached; no more keys required to sign.")
			break
		}
	}

	return suite, pubPoly, sigShares, nil
}
