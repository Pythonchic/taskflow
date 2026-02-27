package email

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

// GenerateCode генерирует 6-значный числовой код
func GenerateCode() string {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		// fallback на время
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	return fmt.Sprintf("%06d", n.Int64())
}
