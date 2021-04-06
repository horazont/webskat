package skat

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrNotEnoughCards = errors.New("fewer cards in deck than requested")
)

type RandomBitstream struct {
	iv           uint64
	counter      uint64
	blockcipher  cipher.Block
	currentBlock io.Reader
}

func SeededBitstream(seed []byte) (*RandomBitstream, error) {
	h := sha512.Sum512_256(seed)
	key := h[:16]
	iv_binary := h[16:24]
	iv := binary.LittleEndian.Uint64(iv_binary)
	blockcipher, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return &RandomBitstream{
		iv:           iv,
		counter:      0,
		blockcipher:  blockcipher,
		currentBlock: nil,
	}, nil
}

func (rb *RandomBitstream) Read(p []byte) (n int, err error) {
	if rb.currentBlock == nil {
		err = rb.generateNewBlock()
		if err != nil {
			return 0, err
		}
	}
	n, err = rb.currentBlock.Read(p)
	if n == 0 {
		rb.currentBlock = nil
		err = rb.generateNewBlock()
		if err != nil {
			return 0, err
		}
		n, err = rb.currentBlock.Read(p)
		if err != nil && n == 0 {
			panic("unexpected read error from fresh block")
		}
		return n, nil
	}
	return n, nil
}

func (rb *RandomBitstream) generateNewBlock() (err error) {
	if rb.counter == 0xffffffffffffffff {
		return io.EOF
	}
	rb.counter = rb.counter + 1
	buf := make([]byte, rb.blockcipher.BlockSize())
	binary.LittleEndian.PutUint64(buf[:8], rb.counter)
	binary.LittleEndian.PutUint64(buf[8:16], rb.iv)
	rb.blockcipher.Encrypt(buf, buf)
	rb.currentBlock = bytes.NewReader(buf)
	return nil
}

type DeterministicDealer struct {
	bitstream io.Reader
}

type Dealer interface {
	DrawCards(cards []Card, ncards uint8) (remainingCards []Card, drawnCards []Card, err error)
}

func getModfactorFromMax(nmax uint8) uint8 {
	if nmax >= 128 {
		return 0
	} else if nmax >= 64 {
		return 128
	} else if nmax >= 32 {
		return 64
	} else if nmax >= 16 {
		return 32
	} else if nmax >= 8 {
		return 16
	} else if nmax >= 4 {
		return 8
	} else if nmax >= 2 {
		return 4
	} else if nmax >= 1 {
		return 2
	} else {
		return 1
	}
}

func PullUint8FromBytesExact(r io.Reader, nmax uint8) (uint8, error) {
	if nmax == 0 {
		return 0, nil
	}
	buf := make([]byte, 1)
	mod := getModfactorFromMax(nmax)
	for {
		nread, err := r.Read(buf)
		if nread != 1 {
			return 0, err
		}
		val := buf[0] % mod
		if val <= nmax {
			return val, nil
		}
	}
}

func NewDealer(seed []byte) (Dealer, error) {
	var bitstream io.Reader
	bitstream, err := SeededBitstream(seed)
	if err != nil {
		return nil, err
	}
	return &DeterministicDealer{bitstream}, nil
}

func (d *DeterministicDealer) DrawCards(cards []Card, ncards uint8) (remainingCards []Card, drawnCards []Card, err error) {
	if len(cards) > 255 {
		panic("too many cards in deck")
	}

	navailable := uint8(len(cards))
	if navailable < ncards {
		return nil, nil, ErrNotEnoughCards
	}

	remainingCards = make([]Card, navailable-ncards)
	drawnCards = make([]Card, ncards)

	for i := uint8(0); i < ncards; i = i + 1 {
		selectedIndex, err := PullUint8FromBytesExact(d.bitstream, navailable-1)
		if err != nil {
			return nil, nil, err
		}
		drawnCards[i] = cards[selectedIndex]
		cards = append(cards[:selectedIndex], cards[selectedIndex+1:]...)
		navailable = navailable - 1
	}

	copy(remainingCards, cards)

	return remainingCards, drawnCards, err
}
