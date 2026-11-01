package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"homework05/bindings/counter"
	"homework05/internal/ethx"
)

// 任务 2.3：使用 abigen 生成的绑定代码与已部署的 Counter 合约交互。
//
// 用法：
//
//	go run ./cmd/counter get
//	go run ./cmd/counter increment
//	go run ./cmd/counter set 42
//	go run ./cmd/counter events
func main() {
	ethx.LoadDotEnv()

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	action := os.Args[1]
	switch action {
	case "get", "increment", "set", "events":
	case "-h", "--help", "help":
		usage()
		return
	default:
		fmt.Fprintf(os.Stderr, "未知操作: %s\n\n", action)
		usage()
		os.Exit(2)
	}

	fs := flag.NewFlagSet(action, flag.ExitOnError)
	addressFlag := fs.String("address", "", "Counter 合约地址；不填则读取 deployments/counter-<chainID>.json")
	dir := fs.String("dir", "deployments", "部署记录目录")
	timeout := fs.Duration("timeout", 3*time.Minute, "等待交易确认的最长时间")
	fromFlag := fs.Int64("from", -1, "events 的起始区块；默认使用部署区块")
	toFlag := fs.Int64("to", -1, "events 的结束区块；默认最新区块")
	fs.Usage = usage
	if err := fs.Parse(os.Args[2:]); err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, chainID, rawURL, err := ethx.Dial(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	address, deployment := resolveAddress(chainID.Int64(), *dir, *addressFlag)
	fmt.Printf("网络        : chain ID %s\n", chainID.String())
	fmt.Printf("RPC         : %s\n", ethx.MaskURL(rawURL))
	fmt.Printf("合约地址    : %s\n", address.Hex())
	if deployment != nil {
		fmt.Printf("部署交易    : %s（区块 %d）\n", deployment.TxHash, deployment.BlockNumber)
	}
	fmt.Println()

	code, err := client.CodeAt(ctx, address, nil)
	if err != nil {
		log.Fatalf("读取合约代码失败: %v", err)
	}
	if len(code) == 0 {
		log.Fatalf("地址 %s 上没有合约代码，请确认地址和网络是否正确", address.Hex())
	}

	contract, err := counter.NewCounter(address, client)
	if err != nil {
		log.Fatalf("创建合约绑定失败: %v", err)
	}

	switch action {
	case "get":
		runGet(ctx, contract)
	case "increment":
		runWrite(ctx, client, chainID, contract, "increment", nil, *timeout)
	case "set":
		if fs.NArg() < 1 {
			log.Fatal("用法: go run ./cmd/counter set <非负整数>")
		}
		value, err := ethx.ParseUint256(fs.Arg(0))
		if err != nil {
			log.Fatalf("计数值无效: %v", err)
		}
		runWrite(ctx, client, chainID, contract, "set", value, *timeout)
	case "events":
		start := uint64(0)
		if *fromFlag >= 0 {
			start = uint64(*fromFlag)
		} else if deployment != nil {
			start = deployment.BlockNumber
		}

		var end *uint64
		if *toFlag >= 0 {
			value := uint64(*toFlag)
			end = &value
		}
		runEvents(ctx, contract, start, end)
	}
}

func runGet(ctx context.Context, contract *counter.Counter) {
	count, err := contract.Count(&bind.CallOpts{Context: ctx})
	if err != nil {
		log.Fatalf("调用 count() 失败: %v", err)
	}
	owner, err := contract.Owner(&bind.CallOpts{Context: ctx})
	if err != nil {
		log.Fatalf("调用 owner() 失败: %v", err)
	}

	fmt.Println("=== 合约读取结果 ===")
	fmt.Printf("owner : %s\n", owner.Hex())
	fmt.Printf("count : %s\n", count.String())
}

func runWrite(
	prepCtx context.Context,
	client *ethclient.Client,
	chainID *big.Int,
	contract *counter.Counter,
	action string,
	value *big.Int,
	timeout time.Duration,
) {
	if err := ethx.RequireSepolia(chainID); err != nil {
		log.Fatal(err)
	}

	key, err := ethx.PrivateKeyFromEnv()
	if err != nil {
		log.Fatal(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(key, chainID)
	if err != nil {
		log.Fatalf("创建交易签名器失败: %v", err)
	}
	auth.Context = prepCtx

	var tx *types.Transaction
	switch action {
	case "increment":
		tx, err = contract.Increment(auth)
	case "set":
		tx, err = contract.SetNumber(auth, value)
	default:
		log.Fatalf("未知写操作: %s", action)
	}
	if err != nil {
		log.Fatalf("发送交易失败: %v", err)
	}

	fmt.Printf("交易已发送  : %s\n", tx.Hash().Hex())
	fmt.Printf("区块浏览器  : https://sepolia.etherscan.io/tx/%s\n", tx.Hash().Hex())

	waitCtx, waitCancel := context.WithTimeout(context.Background(), timeout)
	defer waitCancel()

	fmt.Printf("\n等待交易确认（最多 %s）...\n", timeout.String())
	receipt, err := ethx.WaitForReceipt(waitCtx, client, tx.Hash())
	if err != nil {
		log.Fatalf("等待确认失败: %v", err)
	}

	status := "失败"
	if receipt.Status == types.ReceiptStatusSuccessful {
		status = "成功"
	}
	fmt.Println("\n=== 交易已确认 ===")
	fmt.Printf("状态        : %s（status=%d）\n", status, receipt.Status)
	fmt.Printf("区块号      : %d\n", receipt.BlockNumber.Uint64())
	fmt.Printf("Gas 使用    : %d\n", receipt.GasUsed)

	printCountChanged(contract, receipt)

	readCtx, readCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer readCancel()
	if count, err := contract.Count(&bind.CallOpts{Context: readCtx}); err == nil {
		fmt.Printf("当前 count  : %s\n", count.String())
	} else {
		log.Printf("提示: 读取 count() 失败: %v", err)
	}

	if receipt.Status != types.ReceiptStatusSuccessful {
		os.Exit(1)
	}
}

func printCountChanged(contract *counter.Counter, receipt *types.Receipt) {
	found := 0
	for _, entry := range receipt.Logs {
		event, err := contract.ParseCountChanged(*entry)
		if err != nil {
			continue
		}
		found++
		fmt.Printf("事件        : CountChanged(operator=%s, old=%s, new=%s)\n",
			event.Operator.Hex(), event.OldValue.String(), event.NewValue.String())
	}
	if found == 0 {
		fmt.Println("事件        : 本次回执中没有解析到 CountChanged")
	}
}

func runEvents(ctx context.Context, contract *counter.Counter, start uint64, end *uint64) {
	fmt.Printf("事件范围    : [%d, %s]\n\n", start, endLabel(end))

	iter, err := contract.FilterCountChanged(&bind.FilterOpts{
		Start:   start,
		End:     end,
		Context: ctx,
	}, nil)
	if err != nil {
		log.Fatalf("查询事件失败: %v", err)
	}
	defer iter.Close()

	count := 0
	for iter.Next() {
		event := iter.Event
		count++
		fmt.Printf("[%d] 区块 %d  tx %s\n", count, event.Raw.BlockNumber, event.Raw.TxHash.Hex())
		fmt.Printf("    operator=%s  old=%s  new=%s\n",
			event.Operator.Hex(), event.OldValue.String(), event.NewValue.String())
	}
	if err := iter.Error(); err != nil {
		log.Fatalf("遍历事件失败: %v", err)
	}
	if count == 0 {
		fmt.Println("（指定范围内没有 CountChanged 事件）")
	}
	fmt.Printf("\n共 %d 个事件\n", count)
}

func resolveAddress(chainID int64, dir, flagAddress string) (common.Address, *ethx.Deployment) {
	if flagAddress != "" {
		if !common.IsHexAddress(flagAddress) {
			log.Fatalf("合约地址格式不正确: %q", flagAddress)
		}
		return common.HexToAddress(flagAddress), nil
	}

	deployment, err := ethx.LoadDeployment(dir, chainID)
	if err != nil {
		log.Fatal(err)
	}
	return common.HexToAddress(deployment.Address), deployment
}

func endLabel(end *uint64) string {
	if end == nil {
		return "latest"
	}
	return fmt.Sprintf("%d", *end)
}

func usage() {
	fmt.Fprintln(os.Stderr, `用法:
  go run ./cmd/counter get       [-address 0x...]
  go run ./cmd/counter increment [-address 0x...]
  go run ./cmd/counter set <值>  [-address 0x...]
  go run ./cmd/counter events    [-from 区块号] [-to 区块号]

不填 -address 时，会读取 deployments/counter-<chainID>.json。`)
}
