package ethx

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Deployment struct {
	ChainID      int64     `json:"chainId"`
	ContractName string    `json:"contract"`
	Address      string    `json:"address"`
	TxHash       string    `json:"txHash"`
	BlockNumber  uint64    `json:"blockNumber"`
	DeployedAt   time.Time `json:"deployedAt"`
}

func SaveDeployment(dir string, d Deployment) (string, error) {
	if d.ChainID == 0 {
		return "", errors.New("chain ID 不能为 0")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("创建目录 %s 失败: %w", dir, err)
	}

	path := deploymentPath(dir, d.ChainID)
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return "", fmt.Errorf("序列化部署记录失败: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return "", fmt.Errorf("写入部署记录 %s 失败: %w", path, err)
	}
	return path, nil
}

func LoadDeployment(dir string, chainID int64) (*Deployment, error) {
	path := deploymentPath(dir, chainID)
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("找不到部署记录 %s（请先运行 cmd/deploy，或用 -address 指定合约地址）", path)
		}
		return nil, fmt.Errorf("读取部署记录 %s 失败: %w", path, err)
	}

	var d Deployment
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, fmt.Errorf("解析部署记录 %s 失败: %w", path, err)
	}
	return &d, nil
}

func deploymentPath(dir string, chainID int64) string {
	return filepath.Join(dir, fmt.Sprintf("counter-%d.json", chainID))
}
