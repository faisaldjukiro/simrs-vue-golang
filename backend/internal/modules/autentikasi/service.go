package autentikasi

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrKredensialTidakValid = errors.New("invalid credentials")
var ErrTokenTidakValid = errors.New("invalid access token")

type LoginResult struct {
	TokenType string    `json:"token_type"`
	Token     string    `json:"access_token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      Pengguna  `json:"user"`
}

type Layanan struct {
	repositori *Repositori
	tokenTTL   time.Duration
}

func NewLayanan(repositori *Repositori, tokenTTL time.Duration) *Layanan {
	return &Layanan{repositori: repositori, tokenTTL: tokenTTL}
}

func (s *Layanan) Login(ctx context.Context, username, password string) (LoginResult, error) {
	username = strings.TrimSpace(username)
	if user, ok, err := s.loginDenganSIMRS(ctx, username, password); ok || err != nil {
		if err != nil {
			return LoginResult{}, err
		}
		return s.buatTokenLogin(ctx, user)
	}

	user, err := s.repositori.CariPenggunaByUsername(ctx, username)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && !user.IsActive) {
		return LoginResult{}, ErrKredensialTidakValid
	}
	if err != nil {
		return LoginResult{}, fmt.Errorf("find user: %w", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return LoginResult{}, ErrKredensialTidakValid
	}

	return s.buatTokenLogin(ctx, user)
}

func (s *Layanan) loginDenganSIMRS(ctx context.Context, username, password string) (Pengguna, bool, error) {
	penggunaSIMRS, err := s.repositori.CariPenggunaSIMRS(ctx, username)
	if errors.Is(err, sql.ErrNoRows) {
		return Pengguna{}, false, nil
	}
	if err != nil {
		return Pengguna{}, false, nil
	}
	if subtle.ConstantTimeCompare([]byte(penggunaSIMRS.Password), []byte(password)) != 1 {
		return Pengguna{}, false, nil
	}

	user, err := s.repositori.CariPenggunaByUsername(ctx, penggunaSIMRS.Username)
	if errors.Is(err, sql.ErrNoRows) {
		passwordAcak, err := generateToken()
		if err != nil {
			return Pengguna{}, true, err
		}
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(passwordAcak), bcrypt.DefaultCost)
		if err != nil {
			return Pengguna{}, true, fmt.Errorf("hash generated local password: %w", err)
		}
		user, err = s.repositori.SimpanPenggunaDariSIMRS(ctx, penggunaSIMRS.Username, penggunaSIMRS.Nama, string(passwordHash))
		if err != nil {
			return Pengguna{}, true, err
		}
		return user, true, nil
	}
	if err != nil {
		return Pengguna{}, true, fmt.Errorf("find local SIMRS user: %w", err)
	}
	if !user.IsActive {
		return Pengguna{}, true, ErrKredensialTidakValid
	}
	return user, true, nil
}

func (s *Layanan) buatTokenLogin(ctx context.Context, user Pengguna) (LoginResult, error) {
	token, err := generateToken()
	if err != nil {
		return LoginResult{}, err
	}
	expiresAt := time.Now().UTC().Add(s.tokenTTL)
	s.repositori.HapusTokenKedaluwarsa(ctx)
	if err := s.repositori.SimpanToken(ctx, user.ID, hashToken(token), expiresAt); err != nil {
		return LoginResult{}, err
	}

	return LoginResult{
		TokenType: "Bearer",
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

func (s *Layanan) Authenticate(ctx context.Context, token string) (Pengguna, error) {
	if strings.TrimSpace(token) == "" {
		return Pengguna{}, ErrTokenTidakValid
	}
	user, err := s.repositori.CariPenggunaByToken(ctx, hashToken(token))
	if errors.Is(err, sql.ErrNoRows) {
		return Pengguna{}, ErrTokenTidakValid
	}
	if err != nil {
		return Pengguna{}, fmt.Errorf("authenticate token: %w", err)
	}
	return user, nil
}

func (s *Layanan) Logout(ctx context.Context, token string) error {
	return s.repositori.HapusToken(ctx, hashToken(token))
}

func generateToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate access token: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
