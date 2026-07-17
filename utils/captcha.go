package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wenlng/go-captcha-assets/resources/imagesv2"
	"github.com/wenlng/go-captcha-assets/resources/tiles"
	"github.com/wenlng/go-captcha/v2/slide"
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

var slideCapt slide.Captcha
var slideOnce sync.Once

// InitSlideCaptcha configures and initializes the go-captcha slide builder
func InitSlideCaptcha() {
	slideOnce.Do(func() {
		builder := slide.NewBuilder()

		imgs, err := imagesv2.GetImages()
		if err != nil {
			panic("failed to load go-captcha images: " + err.Error())
		}
		graphs, err := tiles.GetTiles()
		if err != nil {
			panic("failed to load go-captcha tiles: " + err.Error())
		}

		var slideGraphs = make([]*slide.GraphImage, 0, len(graphs))
		for i := 0; i < len(graphs); i++ {
			g := graphs[i]
			slideGraphs = append(slideGraphs, &slide.GraphImage{
				OverlayImage: g.OverlayImage,
				MaskImage:    g.MaskImage,
				ShadowImage:  g.ShadowImage,
			})
		}

		builder.SetResources(
			slide.WithBackgrounds(imgs),
			slide.WithGraphImages(slideGraphs),
		)

		slideCapt = builder.Make()
	})
}

// SlideCaptchaResult contains the visual slices and target positioning
type SlideCaptchaResult struct {
	ID    string `json:"id"`
	Image string `json:"image"`
	Thumb string `json:"thumb"`
	Y     int    `json:"y"`
	W     int    `json:"w"`
	H     int    `json:"h"`
}

// GenerateSlideCaptcha generates a new puzzle slide challenge
func GenerateSlideCaptcha() (*SlideCaptchaResult, error) {
	InitSlideCaptcha()

	captData, err := slideCapt.Generate()
	if err != nil {
		return nil, err
	}

	blockData := captData.GetData()
	targetX := blockData.X
	targetY := blockData.Y

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, err
	}
	id := hex.EncodeToString(idBytes)

	// Keep correct target coordinate in store with a 5-minute expiry
	Store.Set(id, fmt.Sprintf("%d,%d", targetX, targetY), 5*time.Minute)

	masterBase64, err := captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, err
	}

	tileBase64, err := captData.GetTileImage().ToBase64()
	if err != nil {
		return nil, err
	}

	return &SlideCaptchaResult{
		ID:    id,
		Image: masterBase64,
		Thumb: tileBase64,
		Y:     blockData.DY,
		W:     blockData.Width,
		H:     blockData.Height,
	}, nil
}

// VerifySlideCaptcha validates if the user's sliding offset falls within the allowed error margin
func VerifySlideCaptcha(id string, sx, sy int) bool {
	expectedCoordsStr, ok := Store.GetAndRemove(id)
	if !ok {
		return false
	}

	var expectedDX, expectedDY int
	if _, err := fmt.Sscanf(expectedCoordsStr, "%d,%d", &expectedDX, &expectedDY); err != nil {
		return false
	}

	// Verify coordinates with an allowable error margin of 5 pixels
	return slide.Validate(sx, sy, expectedDX, expectedDY, 5)
}
