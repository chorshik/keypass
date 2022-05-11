package pass

import (
	"bytes"
	crand "crypto/rand"
	"fmt"
	"math/big"
	rand "math/rand"
	"os"
	"time"
)

const (
	numbers = "0123456789"
	upper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower   = "abcdefghijklmnopqrstuvwxyz"
	syms    = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

func init() {
	// seed math/rand in case we have to fall back to using it
	rand.Seed(time.Now().Unix() + int64(os.Getpid()+os.Getppid()))
}

// GeneratePassword ....
func GeneratePassword(length int, symbols bool) []byte {
	chars := numbers + upper + lower
	if symbols {
		chars += syms
	}
	if c := os.Getenv("KEYPASS_CHARACTER_SET"); c != "" {
		chars = c
	}
	return []byte(GeneratePasswordCharset(length, chars))
}

// GeneratePasswordCharset ...
func GeneratePasswordCharset(length int, chars string) string {
	pw := &bytes.Buffer{}
	for pw.Len() < length {
		_ = pw.WriteByte(chars[randomInteger(len(chars))])
	}

	return pw.String()
}

func randomInteger(max int) int {
	i, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err == nil {
		return int(i.Int64())
	}
	fmt.Println("WARNING: No crypto/rand available. Falling back to PRNG")
	return rand.Intn(max)
}
