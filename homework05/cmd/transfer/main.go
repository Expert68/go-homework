package main

import (
	"context"
	"flag"
	"fmt"
	"homework05/internal/ethx"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

func main() {
	ethx.LoadDotEnv()

	toHex := flag.String("to", "", "to")
	amountETH := flag.String("amount", "", "转账金额，单位 ETH（必填，例如 0.001）")
	timeout := flag.Duration("timeout", 3*time.Minute, "等待交易确认的最长时间")
	noWait := flag.Bool("no-wait", false, "发送后不等待确认")

	flag.Parse()

	if *toHex == "" || *amountETH == "" {
		flag.Usage()
		log.Fatal("必须同时提供 -to 和 -amount")
	}
	if !common.IsHexAddress(*toHex) {
		log.Fatalf("接收方地址格式不正确: %q", *toHex)
	}
	to := common.HexToAddress(*toHex)
	if to == (common.Address{}) {
		log.Fatal("接收方不能是零地址")
	}

	value, err := ethx.ParseEther(*amountETH)
	if err != nil {
		log.Fatalf("金额无效: %v", err)
	}
	if value.Sign() <= 0 {
		log.Fatal("转账金额必须大于 0")
	}

	key, err := ethx.PrivateKeyFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, chainID, rawURL, err := ethx.Dial(ctx)

	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := ethx.RequireSepolia(chainID); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("网络        : Sepolia（chain ID %s）\n", chainID.String())
	fmt.Printf("RPC         : %s\n", ethx.MaskURL(rawURL))

	prepared, err := ethx.PrepareTransaction(ctx, client, key, ethx.TxRequest{
		To:    to,
		Value: value,
	})
	if err != nil {
		log.Fatalf("构造交易失败: %v", err)
	}
	printPlan(prepared)

	signed, err := prepared.Sign()
	if err != nil {
		log.Fatal(err)
	}
	if err := ethx.Broadcast(ctx, client, signed); err != nil {
		log.Fatal(err)
	}

	hash := signed.Hash()
	fmt.Println("\n=== 交易已发送 ===")
	fmt.Printf("交易哈希    : %s\n", hash.Hex())
	fmt.Printf("区块浏览器  : https://sepolia.etherscan.io/tx/%s\n", hash.Hex())

	if *noWait {
		return
	}

	waitCtx, waitCancel := context.WithTimeout(context.Background(), *timeout)
	defer waitCancel()

	fmt.Printf("\n等待交易确认（最多 %s）...\n", timeout.String())
	receipt, err := ethx.WaitForReceipt(waitCtx, client, hash)
	if err != nil {
		log.Fatalf("等待确认失败: %v", err)
	}
	printReceipt(receipt)

	balanceCtx, balanceCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer balanceCancel()
	if balance, err := client.BalanceAt(balanceCtx, prepared.From, nil); err == nil {
		fmt.Printf("发送后余额  : %s ETH\n", ethx.FormatEther(balance))
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		os.Exit(1)
	}

}

func printPlan(p *ethx.PreparedTx) {
	maxGasCost := new(big.Int).Mul(p.GasFeeCap, new(big.Int).SetUint64(p.GasLimit))

	fmt.Println("\n=== 交易预览 ===")
	fmt.Printf("发送方      : %s\n", p.From.Hex())
	fmt.Printf("接收方      : %s\n", p.To.Hex())
	fmt.Printf("金额        : %s ETH\n", ethx.FormatEther(p.Value))
	fmt.Printf("Nonce       : %d\n", p.Nonce)
	fmt.Printf("Gas Limit   : %d\n", p.GasLimit)
	fmt.Printf("Gas Tip Cap : %s gwei\n", ethx.FormatGwei(p.GasTipCap))
	fmt.Printf("Gas Fee Cap : %s gwei\n", ethx.FormatGwei(p.GasFeeCap))
	fmt.Printf("当前余额    : %s ETH\n", ethx.FormatEther(p.Balance))
	fmt.Printf("最大手续费  : %s ETH（Gas Limit × Gas Fee Cap）\n", ethx.FormatEther(maxGasCost))
}

func printReceipt(r *types.Receipt) {
	status := "失败"
	if r.Status == types.ReceiptStatusSuccessful {
		status = "成功"
	}

	fmt.Println("\n=== 交易已确认 ===")
	fmt.Printf("状态        : %s（status=%d）\n", status, r.Status)
	fmt.Printf("区块号      : %d\n", r.BlockNumber.Uint64())
	fmt.Printf("区块哈希    : %s\n", r.BlockHash.Hex())
	fmt.Printf("Gas 使用    : %d\n", r.GasUsed)
	fmt.Printf("日志数量    : %d\n", len(r.Logs))
}
