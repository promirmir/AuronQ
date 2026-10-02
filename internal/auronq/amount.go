package auronq

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

func ParseAmount(s string) (uint64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, errors.New("empty amount")
	}
	if strings.HasPrefix(s, "-") {
		return 0, errors.New("negative amount")
	}
	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return 0, errors.New("invalid amount")
	}
	whole, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}
	if whole > MaxSupplyCoins {
		return 0, errors.New("amount exceeds max supply")
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	if len(frac) > 8 {
		return 0, errors.New("maximum 8 decimal places")
	}
	frac += strings.Repeat("0", 8-len(frac))
	fv := uint64(0)
	if frac != "" {
		fv, err = strconv.ParseUint(frac, 10, 64)
		if err != nil {
			return 0, err
		}
	}
	if whole > (^uint64(0)-fv)/Coin {
		return 0, errors.New("amount overflow")
	}
	v := whole*Coin + fv
	if v > MaxSupplyAtoms {
		return 0, errors.New("amount exceeds max supply")
	}
	return v, nil
}
func FormatAmount(v uint64) string { return fmt.Sprintf("%d.%08d", v/Coin, v%Coin) }