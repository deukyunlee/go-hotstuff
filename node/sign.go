package node

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"go.dedis.ch/kyber/v3/pairing/bn256"
	"go.dedis.ch/kyber/v3/share"
	"go.dedis.ch/kyber/v3/util/random"
	"log"
	"os"
)

type PartialSecret struct {
	Index    int    `json:"Index"`
	ValueHex string `json:"ValueHex"`
}

var suite = bn256.NewSuite()

// n개 노드, Threshold t를 기반으로 NewPriPoly로 랜덤 키 생성
// 각각 secret_i.json 파일에 저장
func GenerateAndSaveSecrets(n, t int) {
	// 1) 랜덤 다항식(Polynomial) 생성
	//    - share.NewPriPoly(suite.G2(), t, ..., random.New()) 로
	//      degree t-1인 다항식을 생성
	polynomial := share.NewPriPoly(suite.G2(), t, nil, random.New())
	log.Printf("Generated random polynomial (degree=%d), threshold=%d\n", t-1, t)

	// 2) n개의 PriShare 생성
	//    - 각 PriShare는 (Index, Value)를 가짐
	shares := polynomial.Shares(n)

	// 3) secret_i.json 형식으로 저장
	for i := 0; i < n; i++ {
		priShare := shares[i]
		filename := fmt.Sprintf("secret_%d.json", i+1) // i+1 => 노드ID

		err := savePriShareToFile(priShare, filename)
		if err != nil {
			log.Fatalf("failed to save share to file: %v", err)
		}

		log.Printf("Saved partial secret for node %d to %s\n", i+1, filename)
	}
	log.Printf("All %d partial secrets saved.\n", n)
}

// PriShare → JSON 변환하여 파일에 저장
func savePriShareToFile(ps *share.PriShare, filename string) error {
	// ps.Value는 kyber.Scalar
	rawBytes, err := ps.V.MarshalBinary()
	if err != nil {
		return fmt.Errorf("failed to marshal scalar: %v", err)
	}

	secretJSON := PartialSecret{
		Index:    ps.I + 1,
		ValueHex: hex.EncodeToString(rawBytes),
	}

	data, err := json.MarshalIndent(secretJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %v", err)
	}

	return os.WriteFile(filename, data, 0600)
}

// 파일에서 JSON → PriShare 복원
func LoadPriShareFromFile(filename string) (*share.PriShare, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var ps PartialSecret
	if err := json.Unmarshal(data, &ps); err != nil {
		return nil, err
	}

	scalar := suite.G2().Scalar()
	rawBytes, err := hex.DecodeString(ps.ValueHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex for ValueHex: %v", err)
	}
	if err := scalar.UnmarshalBinary(rawBytes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal scalar: %v", err)
	}

	return &share.PriShare{
		I: ps.Index,
		V: scalar,
	}, nil
}
