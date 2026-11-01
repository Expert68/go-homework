package ethx

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type TxBackend interface {
	ChainID(ctx context.Context) (*big.Int, error)
	PendingNonceAt(ctx context.Context, account common.Address) (uint64, error)
	SuggestGasTipCap(ctx context.Context) (*big.Int, error)
	SuggestGasPrice(ctx context.Context) (*big.Int, error)
	HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error)
	EstimateGas(ctx context.Context, msg ethereum.CallMsg) (uint64, error)
	BalanceAt(ctx context.Context, account common.Address, blockNumber *big.Int) (*big.Int, error)
	SendTransaction(ctx context.Context, tx *types.Transaction) error
}

type ReceiptBackend interface {
	TransactionReceipt(ctx context.Context, txHash common.Hash) (*types.Receipt, error)
}

type TxRequest struct {
	To       common.Address
	Value    *big.Int
	Data     []byte
	GasLimit uint64
}

type PreparedTx struct {
	Key       *ecdsa.PrivateKey
	From      common.Address
	To        common.Address
	Value     *big.Int
	Data      []byte
	Nonce     uint64
	GasLimit  uint64
	GasTipCap *big.Int
	GasFeeCap *big.Int
	ChainID   *big.Int
	Balance   *big.Int
}

func PrepareTransaction(ctx context.Context, backend TxBackend, key *ecdsa.PrivateKey, req TxRequest) (*PreparedTx, error) {
	if key == nil {
		return nil, errors.New("私钥为空")
	}

	from := AddressOf(key)
	value := new(big.Int)

	if req.Value != nil {
		value.Set(req.Value)
	}

	if value.Sign() < 0 {
		return nil, errors.New("转账金额不能为负数")
	}

	chainID, err := backend.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取链 ID 失败: %w", err)
	}

	nonce, err := backend.PendingNonceAt(ctx, from)
	if err != nil {
		return nil, fmt.Errorf("读取 nonce 失败: %w", err)
	}

	tipCap, err := backend.SuggestGasTipCap(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取建议的 Gas Tip 失败: %w", err)
	}

	if tipCap == nil {
		tipCap = new(big.Int)
	}

	head, err := backend.HeaderByNumber(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("读取最新区块头失败: %w", err)
	}
	if head == nil {
		return nil, errors.New("节点没有返回最新区块头")
	}

	var feeCap *big.Int
	if head.BaseFee != nil {
		// EIP-1559：fee cap = 2 × base fee + tip（留出 6 个区块的 base fee 波动余量）
		feeCap = new(big.Int).Add(new(big.Int).Lsh(new(big.Int).Set(head.BaseFee), 1), tipCap)
	} else {
		// 兼容不支持 EIP-1559 的链：直接使用节点建议的 gas price
		gasPrice, err := backend.SuggestGasPrice(ctx)
		if err != nil {
			return nil, fmt.Errorf("读取建议的 Gas Price 失败: %w", err)
		}
		feeCap = new(big.Int).Set(gasPrice)
		tipCap = new(big.Int).Set(gasPrice)
	}

	if feeCap.Cmp(tipCap) < 0 {
		feeCap.Set(tipCap)
	}

	gasLimit := req.GasLimit
	if gasLimit == 0 {
		if len(req.Data) == 0 {
			// 普通 ETH 转账固定消耗 21000 gas
			gasLimit = 21000
		} else {
			estimated, err := backend.EstimateGas(ctx, ethereum.CallMsg{
				From:  from,
				To:    &req.To,
				Value: value,
				Data:  req.Data,
			})
			if err != nil {
				return nil, fmt.Errorf("估算 Gas 失败: %w", err)
			}
			// 加 20% 缓冲，避免合约执行时 gas 不足
			gasLimit = estimated * 120 / 100
		}
	}

	balance, err := backend.BalanceAt(ctx, from, nil)
	if err != nil {
		return nil, fmt.Errorf("查询余额失败: %w", err)
	}

	maxCost := new(big.Int).Add(
		new(big.Int).Set(value),
		new(big.Int).Mul(feeCap, new(big.Int).SetUint64(gasLimit)),
	)
	if balance.Cmp(maxCost) < 0 {
		return nil, fmt.Errorf(
			"余额不足：地址 %s 当前余额 %s ETH，本笔交易最多需要 %s ETH（含 Gas 上限费用）",
			from.Hex(), FormatEther(balance), FormatEther(maxCost),
		)
	}

	return &PreparedTx{
		Key:       key,
		From:      from,
		To:        req.To,
		Value:     value,
		Data:      req.Data,
		Nonce:     nonce,
		GasLimit:  gasLimit,
		GasTipCap: tipCap,
		GasFeeCap: feeCap,
		ChainID:   chainID,
		Balance:   balance,
	}, nil

}

func (p *PreparedTx) Sign() (*types.Transaction, error) {
	if p == nil || p.Key == nil {
		return nil, errors.New("交易或私钥为空")
	}

	to := p.To
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   new(big.Int).Set(p.ChainID),
		Nonce:     p.Nonce,
		GasTipCap: new(big.Int).Set(p.GasTipCap),
		GasFeeCap: new(big.Int).Set(p.GasFeeCap),
		Gas:       p.GasLimit,
		To:        &to,
		Value:     new(big.Int).Set(p.Value),
		Data:      p.Data,
	})

	signed, err := types.SignTx(tx, types.NewLondonSigner(p.ChainID), p.Key)
	if err != nil {
		return nil, fmt.Errorf("签名交易失败: %w", err)
	}
	return signed, nil
}

func Broadcast(ctx context.Context, backend TxBackend, tx *types.Transaction) error {
	if tx == nil {
		return errors.New("交易为空")
	}
	if err := backend.SendTransaction(ctx, tx); err != nil {
		return fmt.Errorf("广播交易失败: %w", err)
	}
	return nil
}

func WaitForReceipt(ctx context.Context, backend ReceiptBackend, hash common.Hash) (*types.Receipt, error) {
	return WaitForReceiptWithInterval(ctx, backend, hash, time.Second)
}

func WaitForReceiptWithInterval(ctx context.Context, backend ReceiptBackend, hash common.Hash, interval time.Duration) (*types.Receipt, error) {
	if interval <= 0 {
		interval = time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastErr error
	for {
		receipt, err := backend.TransactionReceipt(ctx, hash)
		if err == nil {
			return receipt, nil
		}
		// 交易还在 pending 时返回 ethereum.NotFound；部分节点在区块刚生成、
		// 索引尚未完成时会返回 "transaction indexing is in progress" 之类的
		// 临时错误。这两类情况都应该继续轮询，而不是立即失败。
		if !errors.Is(err, ethereum.NotFound) {
			lastErr = err
		}

		select {
		case <-ctx.Done():
			if lastErr != nil {
				return nil, fmt.Errorf("等待交易 %s 确认超时（最后一次错误: %v）: %w", hash.Hex(), lastErr, ctx.Err())
			}
			return nil, fmt.Errorf("等待交易 %s 确认超时: %w", hash.Hex(), ctx.Err())
		case <-ticker.C:
		}
	}
}

func SenderOf(tx *types.Transaction, chainID *big.Int) (common.Address, error) {
	return types.Sender(types.LatestSignerForChainID(chainID), tx)
}
