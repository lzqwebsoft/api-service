package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"
)

type captchaData struct {
	code      string
	expiresAt time.Time
}

// CaptchaStore represents a thread-safe registry to store and verify Captchas
type CaptchaStore struct {
	sync.RWMutex
	store map[string]captchaData
}

// Store is the global instance of the CaptchaStore
var Store = &CaptchaStore{
	store: make(map[string]captchaData),
}

// Set stores a Captcha ID with its expected code and TTL
func (s *CaptchaStore) Set(id, code string, ttl time.Duration) {
	s.Lock()
	defer s.Unlock()
	s.store[id] = captchaData{
		code:      strings.ToLower(code),
		expiresAt: time.Now().Add(ttl),
	}
}

// GetAndRemove fetches a Captcha code and deletes it immediately to prevent replay attacks
func (s *CaptchaStore) GetAndRemove(id string) (string, bool) {
	s.Lock()
	defer s.Unlock()
	data, exists := s.store[id]
	if !exists {
		return "", false
	}
	delete(s.store, id) // Single-use token enforcement
	if time.Now().After(data.expiresAt) {
		return "", false
	}
	return data.code, true
}

// GenerateCaptcha creates a random 4-digit code, saves it to store, and returns SVG markup
func GenerateCaptcha() (string, string, error) {
	// Generate random 4-digit numeric code
	digits := "0123456789"
	codeBytes := make([]byte, 4)
	for i := 0; i < 4; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", "", err
		}
		codeBytes[i] = digits[idx.Int64()]
	}
	code := string(codeBytes)

	// Generate a unique ID (UUID style)
	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return "", "", err
	}
	id := hex.EncodeToString(idBytes)

	// Register with a 5-minute expiry window
	Store.Set(id, code, 5*time.Minute)

	// SVG Dimensions
	width := 120
	height := 42

	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, width, height, width, height)
	// Dark UI-compliant Background
	svg += `<rect width="100%" height="100%" fill="#1f1f2e" rx="4"/>`

	// Distortive lines for noise
	for i := 0; i < 4; i++ {
		x1, _ := rand.Int(rand.Reader, big.NewInt(int64(width)))
		y1, _ := rand.Int(rand.Reader, big.NewInt(int64(height)))
		x2, _ := rand.Int(rand.Reader, big.NewInt(int64(width)))
		y2, _ := rand.Int(rand.Reader, big.NewInt(int64(height)))
		svg += fmt.Sprintf(`<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="#3b3b5c" stroke-width="1.5"/>`, x1.Int64(), y1.Int64(), x2.Int64(), y2.Int64())
	}

	// Dots for noise
	for i := 0; i < 20; i++ {
		cx, _ := rand.Int(rand.Reader, big.NewInt(int64(width)))
		cy, _ := rand.Int(rand.Reader, big.NewInt(int64(height)))
		r, _ := rand.Int(rand.Reader, big.NewInt(2))
		svg += fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="#4d4d7a" opacity="0.6"/>`, cx.Int64(), cy.Int64(), r.Int64()+1)
	}

	// Dynamic text placement and rotation
	colors := []string{"#ff5555", "#50fa7b", "#f1fa8c", "#bd93f9", "#ff79c6", "#8be9fd"}
	for i, char := range code {
		cIdx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(colors))))
		color := colors[cIdx.Int64()]

		fontSize := 22 + i%2*2
		x := 15 + i*24
		y := 28

		angle, _ := rand.Int(rand.Reader, big.NewInt(30))
		rot := angle.Int64() - 15 // Rotate between -15 to +15 degrees

		svg += fmt.Sprintf(`<text x="%d" y="%d" font-family="Arial, sans-serif" font-size="%d" font-weight="bold" fill="%s" transform="rotate(%d %d %d)">%c</text>`, x, y, fontSize, color, rot, x, y, char)
	}

	svg += `</svg>`
	return id, svg, nil
}
