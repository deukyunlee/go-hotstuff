package signature

import (
	"go.dedis.ch/kyber/v4/pairing/bn256"
	"go.dedis.ch/kyber/v4/share"
	"go.dedis.ch/kyber/v4/sign/bls"
	"go.dedis.ch/kyber/v4/sign/tbls"
	"math"
)

func VerifyMessage(n int, msg []byte, suite *bn256.Suite, pubPoly *share.PubPoly, sigShares [][]byte) error {
	t := int(math.Ceil(float64(n) / 3 * 2))

	sig, err := tbls.Recover(suite, pubPoly, msg, sigShares, t, n)
	if err != nil {
		logger.Errorf("Error occured while recovering: %s", err)
		return err
	}

	err = bls.Verify(suite, pubPoly.Commit(), msg, sig)
	if err != nil {
		logger.Errorf("Error occured while verifying: %s", err)
		return err
	}

	// sig: 전체 서명
	logger.Infof("Msg [%s] Verified", string(msg))
	return nil
}
