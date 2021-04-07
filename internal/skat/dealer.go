package skat

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"errors"
	"io"
)

var (
	ErrNotEnoughCards      = errors.New("fewer cards in deck than requested")
	ErrIncorrectSeedLength = errors.New("incorrect seed length")
	ErrTooManyItems        = errors.New("shuffle is restricted to 256 items at most")
)

type NullReader struct{}

func (r *NullReader) Read(dst []byte) (n int, err error) {
	for i := range dst {
		dst[i] = 0
	}
	return len(dst), nil
}

// Create an io.Reader which generates pseudorandom bytes from a given seed
//
// The seed must be exactly 32 bytes long.
func SeededBitstream(seed []byte) (io.Reader, error) {
	if len(seed) != 32 {
		return nil, ErrIncorrectSeedLength
	}
	blockCipher, err := aes.NewCipher(seed[:16])
	if err != nil {
		return nil, err
	}
	return &cipher.StreamReader{
		S: cipher.NewCTR(
			blockCipher,
			seed[16:],
		),
		R: &NullReader{},
	}, nil
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

// Shuffle an array in-place
//
// Currently, only up to 255 items are supported.
func Shuffle(r io.Reader, data *CardSet) error {
	if len(*data) > 256 {
		return ErrTooManyItems
	}

	nitems := uint8(len(*data))

	for i_ := range *data {
		i := uint8(i_)
		remaining := nitems - i
		chosen, err := PullUint8FromBytesExact(r, remaining-1)
		if err != nil {
			return err
		}
		buf := (*data)[i]
		(*data)[i] = (*data)[chosen]
		(*data)[chosen] = buf
	}

	return nil
}

// Shuffle a deck of cards with an arbitrarily-sized seed
//
// Note: To ensure that the cards are shuffled with all fairness, the seed
// should contain at least 120 bits of entropy, the more the better.
func ShuffleDeckWithSeed(seed []byte, data *CardSet) error {
	hashedSeed := sha512.Sum512(seed)
	bitstream1, err := SeededBitstream(hashedSeed[:32])
	if err != nil {
		return err
	}
	bitstream2, err := SeededBitstream(hashedSeed[32:])
	if err != nil {
		return err
	}

	// Why two shuffles with two different bitstreams? The reason is that
	// AES-128, which is used to back the stream cipher for the csprng, may
	// not be able to generate all hands. This may be my lack of
	// understanding.
	//
	// Here is my reasoning:
	// - A deck of 32 cards can be shuffled in 32! (factorial) ways.
	// - The entropy "contained" in a shuffle is log2(32!) which is
	//   approximately 118 bits
	// - The AES-128 key length is 128 bits
	// - The AES-128 security margin is probably around 100 bits, depending on
	//   the exact scenario and which attacks are possible.
	//
	// Given all this, it is possible that AES-128 does not "transfer" the
	// complete amount of entropy from its key to its output stream. Even if
	// it does, a margin of 10 bits (2^10 = 1024) is not a lot to ensure that
	// the shuffles are unbiased.
	//
	// By hashing the initial seed with SHA512, splitting the hash output in
	// two halves and initializing two AES-128-based stream ciphers with those
	// separate halves, I believe I am increasing the "bandwidth" of entropy
	// (so to speak) from the seed to the shuffled deck and thus increasing
	// the fairness of the shuffle.

	// Shuffle shuffles the array data in-place using the random data sourced
	// from the given io.Reader.
	err = Shuffle(bitstream1, data)
	if err != nil {
		return err
	}
	err = Shuffle(bitstream2, data)
	if err != nil {
		return err
	}

	return nil
}

func DrawCards(deck CardSet, ncards int) (remainingCards CardSet, drawnCards CardSet, err error) {
	if len(deck) < ncards {
		return nil, nil, ErrNotEnoughCards
	}

	remainingCards = deck[ncards:].Copy()
	drawnCards = deck[:ncards].Copy()
	return remainingCards, drawnCards, nil
}
