package main

import (
	"context"
	"flag"
	"fmt"
	"homework05/internal/ethx"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

func main() {
	ethx.LoadDotEnv()
	addressFlag := flag.String("address", "", "要查询的地址；不填则使用 SEPOLIA_PRIVATE_KEY 推导出的地址")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, chainID, rawURL, err := ethx.Dial(ctx)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Close()

	var (
		address common.Address
		source  string
	)

	if *addressFlag != "" {
		if !common.IsHexAddress(*addressFlag) {
			log.Fatalf("地址格式不正确: %q", *addressFlag)
		}
		address = common.HexToAddress(*addressFlag)
		source = "命令行 -address"
	} else {
		key, err := ethx.PrivateKeyFromEnv()
		if err != nil {
			log.Fatalf("没有提供 -address，也没有可用私钥: %v", err)
		}
		address = ethx.AddressOf(key)
		source = "私钥推导"
	}

	balance, err := client.BalanceAt(ctx, address, nil)
	if err != nil {
		log.Fatalf("查询余额失败: %v", err)
	}
	nonce, err := client.PendingNonceAt(ctx, address)
	if err != nil {
		log.Fatalf("查询 nonce 失败: %v", err)
	}
	code, err := client.CodeAt(ctx, address, nil)
	if err != nil {
		log.Fatalf("查询账户代码失败: %v", err)
	}

	kind := "普通账户（EOA）"
	if len(code) > 0 {
		kind = fmt.Sprintf("合约账户（%d 字节代码）", len(code))
	}
	network := "未知网络"
	if chainID.IsInt64() && chainID.Int64() == ethx.SepoliaChainID {
		network = "Sepolia 测试网"
	}

	fmt.Println("=== 账户信息 ===")
	fmt.Printf("网络      : %s（chain ID %s）\n", network, chainID.String())
	fmt.Printf("RPC       : %s\n", ethx.MaskURL(rawURL))
	fmt.Printf("地址      : %s（%s）\n", address.Hex(), source)
	fmt.Printf("类型      : %s\n", kind)
	fmt.Printf("余额      : %s ETH\n", ethx.FormatEther(balance))
	fmt.Printf("Nonce     : %d\n", nonce)

	if balance.Sign() == 0 {
		fmt.Println("\n提示: 余额为 0，可以从 Sepolia 水龙头领取测试 ETH。")
	}
}
