// Package service implements the URL Shortener service.
package service

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"strings"

	"github.com/avalonbits/shortener/storage"
)

// Shortener is the implemention of the URL Shortener service.
type Shortener struct {
	queries     *storage.Queries
	randSource  io.Reader
	existsRetry int
}

// NewShortener creates a Shortener with the needed dependencies.
// queries is the sqlc generated code to access the database.
// randSource is a source of random types. For true randomness, user the Reader from crypt/rand.
// existsRetry is the number of times to generate a new random short name in case it already exists
// in the database. It defaults to 1 if existsRetry < 1.
func NewShortener(queries *storage.Queries, randSource io.Reader, existsRetry int) *Shortener {
	if existsRetry < 1 {
		existsRetry = 1
	}
	return &Shortener{
		queries:     queries,
		randSource:  randSource,
		existsRetry: existsRetry,
	}
}

var (
	ErrEmptyString   = fmt.Errorf("empty strings are not allowed")
	ErrGenerateShort = fmt.Errorf("unable to create short name")
	ErrDatabase      = fmt.Errorf("database access error")
	ErrShortName     = fmt.Errorf("invalid short name")
	ErrExists        = fmt.Errorf("short name already exists")
	ErrNotFound      = fmt.Errorf("short name not found")
	ErrInvalidURL    = fmt.Errorf("invalid url")
	ErrTooLong       = fmt.Errorf("URL is too long")
)

// ShortNameFor stores and returns a short name for the given long string.
func (s *Shortener) ShortNameFor(ctx context.Context, long string) (string, error) {
	long, err := s.Validate(long)
	if err != nil {
		return "", err
	}

	var short string
	err = nil
	for i := 0; i <= s.existsRetry; i++ {
		short, err = GenerateShort(s.randSource)
		if err != nil {
			return "", fmt.Errorf("%w: %v", ErrGenerateShort, err)
		}

		err = s.queries.SetShort(ctx, storage.SetShortParams{
			Short: short,
			Longn: long,
		})

		// If no errors, we are done.
		if err == nil {
			return short, nil
		}

		// We only want to retry if the error is a constraint violation, because
		// that would mean we tried to insert a duplicate short name.
		if storage.IsErrConstraint(err) {
			err = ErrExists
		} else {
			return "", fmt.Errorf("%w: %v", ErrDatabase, err)
		}
	}

	return "", err
}

// Validate checks if the long string is a safe URL that we can store.
func (s *Shortener) Validate(long string) (string, error) {
	long = strings.TrimSpace(long)
	if long == "" {
		return "", ErrEmptyString
	}

	const urlLength = 8 * 1024
	if len(long) > urlLength {
		return "", ErrTooLong
	}

	// TODO(icc): implement URL validation.
	return long, nil
}

// LongFrom returns the long name for the given short, if it was previously stored.
func (s *Shortener) LongFrom(ctx context.Context, short string) (string, error) {
	short = strings.TrimSpace(short)
	if len(short) != 8 {
	}

	long, err := s.queries.GetLong(ctx, short)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("%w: %v", ErrDatabase, err)
	}
	return long, nil
}

// GenerateShort creates a random 8 character string.
// The string as a base64 encoded 6 byte sequence, allowing the creationg of 2⁴⁸ unique strings.
func GenerateShort(randSource io.Reader) (string, error) {
	// Allocate a single buffer that can store the 6 byte long key and the 8 byte encoded version.
	const keyLen = 6
	const base64len = 8
	buf := make([]byte, keyLen+base64len)

	randB := buf[:keyLen]
	n, err := randSource.Read(randB)
	if err != nil {
		return "", fmt.Errorf("error creating short url: %w", err)
	}
	if n != keyLen {
		return "", fmt.Errorf("unable to read data needed to create short url")
	}

	dst := buf[keyLen:]
	base64.URLEncoding.Encode(dst, randB)

	return string(dst), nil
}
