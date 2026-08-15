package bpjs

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

const alfabetURI = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+-$"

// DekripsiResponse membuka response VClaim 2.0 melalui Base64 decode,
// AES-256-CBC decrypt, lalu dekompresi LZString encoded URI component.
func DekripsiResponse(response, consumerID, secretKey, timestamp string) ([]byte, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimSpace(response))
	if err != nil {
		return nil, fmt.Errorf("response bukan Base64: %w", err)
	}
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return nil, fmt.Errorf("panjang ciphertext AES tidak valid")
	}

	hash := sha256.Sum256([]byte(consumerID + secretKey + timestamp))
	block, err := aes.NewCipher(hash[:])
	if err != nil {
		return nil, fmt.Errorf("membuat cipher AES: %w", err)
	}
	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, hash[:aes.BlockSize]).CryptBlocks(plaintext, ciphertext)
	plaintext, err = hapusPaddingPKCS7(plaintext)
	if err != nil {
		return nil, err
	}

	hasil, err := dekompresiLZStringURI(string(plaintext))
	if err != nil {
		return nil, err
	}
	return []byte(hasil), nil
}

func hapusPaddingPKCS7(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("hasil dekripsi kosong")
	}
	padding := int(data[len(data)-1])
	if padding < 1 || padding > aes.BlockSize || padding > len(data) {
		return nil, fmt.Errorf("padding response tidak valid")
	}
	for _, value := range data[len(data)-padding:] {
		if int(value) != padding {
			return nil, fmt.Errorf("padding response tidak konsisten")
		}
	}
	return data[:len(data)-padding], nil
}

type pembacaBitLZ struct {
	nilai     int
	posisi    int
	index     int
	panjang   int
	nilaiPada func(int) (int, error)
}

func (p *pembacaBitLZ) baca(jumlah int) (int, error) {
	bits, power := 0, 1
	maxPower := 1 << jumlah
	for power != maxPower {
		resb := p.nilai & p.posisi
		p.posisi >>= 1
		if p.posisi == 0 {
			p.posisi = 32
			if p.index >= p.panjang {
				return 0, fmt.Errorf("data LZString terpotong")
			}
			nilai, err := p.nilaiPada(p.index)
			if err != nil {
				return 0, err
			}
			p.nilai = nilai
			p.index++
		}
		if resb > 0 {
			bits |= power
		}
		power <<= 1
	}
	return bits, nil
}

func dekompresiLZStringURI(input string) (string, error) {
	input = strings.ReplaceAll(input, " ", "+")
	if input == "" {
		return "", nil
	}
	runes := []rune(input)
	lookup := make(map[rune]int, len([]rune(alfabetURI)))
	for index, karakter := range []rune(alfabetURI) {
		lookup[karakter] = index
	}
	nilaiPada := func(index int) (int, error) {
		nilai, ada := lookup[runes[index]]
		if !ada {
			return 0, fmt.Errorf("karakter LZString tidak valid")
		}
		return nilai, nil
	}
	nilaiAwal, err := nilaiPada(0)
	if err != nil {
		return "", err
	}
	pembaca := &pembacaBitLZ{
		nilai: nilaiAwal, posisi: 32, index: 1, panjang: len(runes), nilaiPada: nilaiPada,
	}

	kamus := []string{"0", "1", "2"}
	jenis, err := pembaca.baca(2)
	if err != nil {
		return "", err
	}
	var karakter string
	switch jenis {
	case 0:
		nilai, err := pembaca.baca(8)
		if err != nil {
			return "", err
		}
		karakter = string(rune(nilai))
	case 1:
		nilai, err := pembaca.baca(16)
		if err != nil {
			return "", err
		}
		karakter = string(rune(nilai))
	case 2:
		return "", nil
	default:
		return "", fmt.Errorf("header LZString tidak valid")
	}

	kamus = append(kamus, karakter)
	w := karakter
	var hasil strings.Builder
	hasil.WriteString(karakter)
	enlargeIn, dictSize, numBits := 4, 4, 3

	for {
		kode, err := pembaca.baca(numBits)
		if err != nil {
			return "", err
		}
		switch kode {
		case 0:
			nilai, err := pembaca.baca(8)
			if err != nil {
				return "", err
			}
			kamus = append(kamus, string(rune(nilai)))
			kode = dictSize
			dictSize++
			enlargeIn--
		case 1:
			nilai, err := pembaca.baca(16)
			if err != nil {
				return "", err
			}
			kamus = append(kamus, string(rune(nilai)))
			kode = dictSize
			dictSize++
			enlargeIn--
		case 2:
			return hasil.String(), nil
		}

		if enlargeIn == 0 {
			enlargeIn = 1 << numBits
			numBits++
		}

		var entry string
		if kode < len(kamus) {
			entry = kamus[kode]
		} else if kode == dictSize {
			entry = w + string([]rune(w)[0])
		} else {
			return "", fmt.Errorf("kamus LZString tidak valid")
		}
		if entry == "" {
			return "", fmt.Errorf("entry LZString kosong")
		}
		hasil.WriteString(entry)
		kamus = append(kamus, w+string([]rune(entry)[0]))
		dictSize++
		enlargeIn--
		w = entry

		if enlargeIn == 0 {
			enlargeIn = 1 << numBits
			numBits++
		}
	}
}
