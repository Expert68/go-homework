package ethx

import (
	"fmt"
	"math/big"
	"strings"
)

var (
	weiPerEther = new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	weiPerGwei  = big.NewInt(1_000_000_000)
)

func ParseEther(s string) (*big.Int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("金额不能为空")
	}
	if strings.HasPrefix(s, "+") {
		s = s[1:]
	}
	if strings.HasPrefix(s, "-") {
		return nil, fmt.Errorf("金额不能为负数: %q", s)
	}
	if strings.Count(s, ".") > 1 {
		return nil, fmt.Errorf("金额格式不正确: %q", s)
	}

	if s == "." {
		return nil, fmt.Errorf("金额格式不正确: %q", s)
	}

	whole, frac, _ := strings.Cut(s, ".")

	if whole == "" {
		whole = "0"
	}

	if !allDigits(whole) || !allDigits(frac) {
		return nil, fmt.Errorf("金额格式不正确: %q", s)
	}
	if len(frac) > 18 {
		return nil, fmt.Errorf("金额最多支持 18 位小数: %q", s)
	}

	frac += strings.Repeat("0", 18-len(frac))
	wei, ok := new(big.Int).SetString(whole+frac, 10)
	if !ok {
		return nil, fmt.Errorf("金额格式不正确: %q", s)
	}
	return wei, nil
}

func ParseUint256(s string) (*big.Int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("数值不能为空")
	}

	if strings.HasPrefix(s, "-") {
		return nil, fmt.Errorf("数值不能为负数: %q", s)
	}
	n, ok := new(big.Int).SetString(s, 10)

	if !ok {
		return nil, fmt.Errorf("数值格式不正确: %q", s)
	}
	if n.BitLen() > 256 {
		return nil, fmt.Errorf("数值超过uint256范围： %q", s)
	}
	return n, nil
}

func FormatEther(wei *big.Int) string {
	if wei == nil {
		return "0"
	}
	if wei.Sign() < 0 {
		return "-" + FormatEther(new(big.Int).Neg(wei))
	}

	quotient, remainder := new(big.Int).QuoRem(wei, weiPerEther, new(big.Int))
	if remainder.Sign() == 0 {
		return quotient.String()
	}

	frac := remainder.String()
	frac = strings.Repeat("0", 18-len(frac)) + frac
	frac = strings.TrimRight(frac, "0")
	return quotient.String() + "." + frac
}

func FormatGwei(wei *big.Int) string {
	if wei == nil {
		return "0"
	}
	if wei.Sign() < 0 {
		return "-" + FormatGwei(new(big.Int).Neg(wei))
	}

	quotient, remainder := new(big.Int).QuoRem(wei, weiPerGwei, new(big.Int))
	if remainder.Sign() == 0 {
		return quotient.String()
	}

	frac := remainder.String()
	frac = strings.Repeat("0", 9-len(frac)) + frac
	frac = strings.TrimRight(frac, "0")
	return quotient.String() + "." + frac
}

func allDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
