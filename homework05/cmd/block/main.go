package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/core/types"

	"homework05/internal/ethx"
)

func main() {
	ethx.LoadDotEnv()

	numberArg := flag.String("number", "latest", "要查询的区块号（十进制），或 latest")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, chainID, rawURL, err := ethx.Dial(ctx)
	if err != nil {
		log.Fatalf("错误: %v", err)
	}
	defer client.Close()

	var number *big.Int
	if !strings.EqualFold(strings.TrimSpace(*numberArg), "latest") {
		n, err := strconv.ParseUint(strings.TrimSpace(*numberArg), 10, 64)
		if err != nil {
			log.Fatalf("无效的区块号 %q：请输入十进制数字或 latest", *numberArg)
		}
		number = new(big.Int).SetUint64(n)
	}

	block, err := client.BlockByNumber(ctx, number)
	if err != nil {
		log.Fatalf("查询区块失败: %v", err)
	}

	latestHeader, err := client.HeaderByNumber(ctx, nil)
	if err != nil {
		log.Printf("提示: 查询最新区块高度失败（不影响本次区块信息）: %v", err)
		latestHeader = nil
	}

	printBlock(rawURL, chainID, block, latestHeader)
}

func printBlock(rawURL string, chainID *big.Int, block *types.Block, latest *types.Header) {
	network := "未知网络"
	if chainID.IsInt64() && chainID.Int64() == ethx.SepoliaChainID {
		network = "Sepolia 测试网"
	}

	fmt.Println("========================================")
	fmt.Println(" 以太坊区块信息")
	fmt.Println("========================================")
	fmt.Printf("网络          : %s（chain ID %s）\n", network, chainID.String())
	fmt.Printf("RPC           : %s\n", ethx.MaskURL(rawURL))
	fmt.Printf("区块号        : %d\n", block.NumberU64())
	fmt.Printf("区块哈希      : %s\n", block.Hash().Hex())
	fmt.Printf("父区块哈希    : %s\n", block.ParentHash().Hex())

	blockTime := time.Unix(int64(block.Time()), 0)
	fmt.Printf("时间戳(UTC)   : %s\n", blockTime.UTC().Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("时间戳(UTC+8) : %s\n", blockTime.In(utc8()).Format("2006-01-02 15:04:05"))

	fmt.Printf("交易数量      : %d\n", len(block.Transactions()))
	if block.GasLimit() > 0 {
		usage := float64(block.GasUsed()) / float64(block.GasLimit()) * 100
		fmt.Printf("Gas 使用/上限 : %d / %d（%.2f%%）\n", block.GasUsed(), block.GasLimit(), usage)
	}
	if baseFee := block.BaseFee(); baseFee != nil {
		fmt.Printf("Base Fee      : %s gwei\n", ethx.FormatGwei(baseFee))
	}
	fmt.Printf("Coinbase      : %s\n", block.Coinbase().Hex())
	fmt.Printf("状态根        : %s\n", block.Root().Hex())

	if latest != nil && latest.Number != nil && latest.Number.Uint64() >= block.NumberU64() {
		fmt.Printf("确认数        : %d\n", latest.Number.Uint64()-block.NumberU64())
	}

	printTransactions(block, chainID)
}

func printTransactions(block *types.Block, chainID *big.Int) {
	txs := block.Transactions()
	if len(txs) == 0 {
		fmt.Println("\n（该区块没有交易）")
		return
	}

	const maxPrint = 5
	limit := len(txs)
	if limit > maxPrint {
		limit = maxPrint
	}

	fmt.Printf("\n--- 前 %d 笔交易 ---\n", limit)
	for i := 0; i < limit; i++ {
		tx := txs[i]

		from := "（无法恢复发送方）"
		if sender, err := ethx.SenderOf(tx, chainID); err == nil {
			from = sender.Hex()
		}

		to := "（合约创建）"
		if tx.To() != nil {
			to = tx.To().Hex()
		}

		fmt.Printf("[%d] %s\n", i+1, tx.Hash().Hex())
		fmt.Printf("    from=%s\n", from)
		fmt.Printf("    to=%s\n", to)
		fmt.Printf("    value=%s ETH  type=%d  nonce=%d\n",
			ethx.FormatEther(tx.Value()), tx.Type(), tx.Nonce())
	}
	if len(txs) > limit {
		fmt.Printf("... 其余 %d 笔交易省略\n", len(txs)-limit)
	}
}

func utc8() *time.Location {
	return time.FixedZone("UTC+8", 8*60*60)
}
