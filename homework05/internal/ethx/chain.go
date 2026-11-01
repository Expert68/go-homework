package ethx

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
)

const SepoliaChainID int64 = 11155111

func RPCURL() (string, error) {
	for _, name := range []string{"SEPOLIA_RPC_URL", "ETH_RPC_URL", "RPC_URL"} {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value, nil
		}
	}
	return "", errors.New("未找到 RPC 地址： 请设置环境变量 SEPOLIA_RPC_URL(例如: https://sepolia.infura.io/v3/<your-key>)")
}

func Dial(ctx context.Context) (*ethclient.Client, *big.Int, string, error) {
	rawURL, err := RPCURL()
	if err != nil {
		return nil, nil, "", err
	}
	client, err := ethclient.Dial(rawURL)
	if err != nil {
		return nil, nil, "", fmt.Errorf("连接RPC失败(%s): %w", MaskURL(rawURL), err)
	}

	chainID, err := client.ChainID(ctx)
	if err != nil {
		client.Close()
		return nil, nil, "", fmt.Errorf("读取链失败: %w", err)
	}
	return client, chainID, rawURL, nil
}

func RequireSepolia(ChainID *big.Int) error {
	if ChainID == nil {
		return errors.New("链ID为空")
	}
	if ChainID.Cmp(big.NewInt(SepoliaChainID)) != 0 {
		return fmt.Errorf("当前链ID为 %s,不是Sepolia (%d);为避免误操作，写操作已中止", ChainID.String(), SepoliaChainID)
	}
	return nil
}

func MaskURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return raw
	}

	segments := strings.Split(strings.Trim(u.Path, "/"), "/")
	for i, segment := range segments {
		if len(segment) >= 16 {
			segments[i] = "***"
		}
	}
	if len(segments) > 0 && segments[0] != "" {
		u.Path = "/" + strings.Join(segments, "/")
	} else {
		u.Path = ""
	}
	u.RawQuery = ""
	u.User = nil
	return u.String()
}
