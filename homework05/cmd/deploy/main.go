package main

import (
	"context"
	"flag"
	"fmt"
	"homework05/bindings/counter"
	"homework05/internal/ethx"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
)

func main() {
	ethx.LoadDotEnv()

	dir := flag.String("dir", "deployments", "部署记录保存目录")
	timeout := flag.Duration("timeout", 3*time.Minute, "等待部署交易确认的最长时间")
	flag.Parse()

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

	from := ethx.AddressOf(key)
	balance, err := client.BalanceAt(ctx, from, nil)
	if err != nil {
		log.Fatalf("查询部署账户余额失败: %v", err)
	}

	fmt.Println("=== 部署 Counter 合约 ===")
	fmt.Printf("网络        : Sepolia（chain ID %s）\n", chainID.String())
	fmt.Printf("RPC         : %s\n", ethx.MaskURL(rawURL))
	fmt.Printf("部署账户    : %s\n", from.Hex())
	fmt.Printf("账户余额    : %s ETH\n", ethx.FormatEther(balance))
	if balance.Sign() == 0 {
		log.Fatal("账户余额为 0，请先从 Sepolia 水龙头领取测试 ETH")
	}

	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		log.Fatalf("创建交易签名器失败: %v", err)
	}
	auth.Context = ctx

	fmt.Println("\n正在发送部署交易...")
	address, tx, contract, err := counter.DeployCounter(auth, client)

	if err != nil {
		log.Fatalf("部署失败: %v", err)
	}
	fmt.Printf("合约地址    : %s\n", address.Hex())
	fmt.Printf("部署交易    : %s\n", tx.Hash().Hex())
	fmt.Printf("区块浏览器  : https://sepolia.etherscan.io/tx/%s\n", tx.Hash().Hex())

	waitCtx, waitCancel := context.WithTimeout(context.Background(), *timeout)
	defer waitCancel()

	fmt.Printf("\n等待部署确认（最多 %s）...\n", timeout.String())
	receipt, err := ethx.WaitForReceipt(waitCtx, client, tx.Hash())
	if err != nil {
		log.Fatalf("等待部署确认失败: %v", err)
	}

	status := "失败"
	if receipt.Status == 1 {
		status = "成功"
	}

	fmt.Println("\n=== 部署结果 ===")
	fmt.Printf("状态        : %s（status=%d）\n", status, receipt.Status)
	fmt.Printf("区块号      : %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("Gas 使用    : %d\n", receipt.GasUsed)
	fmt.Printf("合约地址    : %s\n", address.Hex())
	fmt.Printf("合约浏览器  : https://sepolia.etherscan.io/address/%s\n", address.Hex())

	if receipt.Status != 1 {
		log.Fatal("部署交易执行失败，请到区块浏览器查看失败原因")
	}

	if n, err := contract.Count(&bind.CallOpts{Context: waitCtx}); err == nil {
		fmt.Printf("初始 count  : %s\n", n.String())
	}

	deployment := ethx.Deployment{
		ChainID:      chainID.Int64(),
		ContractName: "Counter",
		Address:      address.Hex(),
		TxHash:       tx.Hash().Hex(),
		BlockNumber:  receipt.BlockNumber.Uint64(),
		DeployedAt:   time.Now().UTC(),
	}
	path, err := ethx.SaveDeployment(*dir, deployment)
	if err != nil {
		log.Printf("警告: 保存部署记录失败: %v", err)
	} else {
		fmt.Printf("部署记录    : %s\n", path)
	}
}
