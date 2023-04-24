package evendeep

import (
	"math/big"
	mrand "math/rand"
	"sync"
	"time"
)

var (
	Randtool = &randomizer{}

	hundred    = big.NewInt(100)                                         //nolint:unused,deadcode,varcheck //test
	seededRand = mrand.New(mrand.NewSource(time.Now().UTC().UnixNano())) //nolint:gosec //G404: Use of weak random number generator (math/rand instead of crypto/rand)
	mu         sync.Mutex
)

// var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
const (
	// Alphabets gets the a to z and A to Z
	Alphabets = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Digits gets 0 to 9
	Digits = "0123456789"
	// AlphabetNumerics gets Alphabets and Digits
	AlphabetNumerics = Alphabets + Digits
	// Symbols gets the ascii symbols
	Symbols = "~!@#$%^&*()-_+={}[]\\|<,>.?/\"';:`"
	// ASCII gets the ascii characters
	ASCII = AlphabetNumerics + Symbols
)

type randomizer struct {
	lastErr error //nolint:unused,structcheck //usable
}

func (r *randomizer) Next() int {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Int()
}

func (r *randomizer) NextIn(max int) int {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Intn(max)
}

func (r *randomizer) inRange(min, max int) int {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Intn(max-min) + min
}

func (r *randomizer) NextInRange(min, max int) int { return r.inRange(min, max) }

func (r *randomizer) NextInt63n(n int64) int64 {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Int63n(n)
}

func (r *randomizer) NextIntn(n int) int {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Intn(n)
}

func (r *randomizer) NextFloat64() float64 {
	mu.Lock()
	defer mu.Unlock()
	return seededRand.Float64()
}

// NextStringSimple returns a random string with specified length 'n', just in A..Z
func (r *randomizer) NextStringSimple(n int) string {
	b := make([]byte, n)
	for i := 0; i < n; i++ {
		n := seededRand.Intn(90-65) + 65
		b[i] = byte(n) // 'a' .. 'z'
	}
	return string(b)
}
