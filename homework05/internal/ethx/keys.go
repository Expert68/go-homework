package ethx

import (
	"fmt"
	"os"
	"strings"

	"crypto/ecdsa"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func PrivateKeyFromEnv() (*ecdsa.PrivateKey, error) {
	for _, name := range []string{"SEPOLIA_PRIVATE_KEY", "SENDER_PRIVATE_KEY", "PRIVATE_KEY"} {
		value := strings.TrimSpace(os.Getenv(name))

		if value == "" {
			continue
		}
		return ParsePrivateKey(value)
	}
	return nil, fmt.Errorf("未找到私钥：请设置环境变量 SEPOLIA_PRIVATE_KEY（仅使用测试网账户）")
}

func ParsePrivateKey(hexKey string) (*ecdsa.PrivateKey, error) {
	s := strings.TrimSpace(hexKey)
	s = strings.TrimPrefix(s, "0x")
	s = strings.TrimPrefix(s, "0X")

	if len(s) != 64 {
		return nil, fmt.Errorf("私钥必须是 32 字节十六进制（64 个字符），当前长度为 %d", len(s))
	}

	key, err := crypto.HexToECDSA(s)

	if err != nil {
		return nil, fmt.Errorf("解析私钥失败: %w", err)
	}
	return key, nil
}

func AddressOf(key *ecdsa.PrivateKey) common.Address { return crypto.PubkeyToAddress(key.PublicKey) }
