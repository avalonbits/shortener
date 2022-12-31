package service_test

import (
	"context"
	"encoding/base64"
	"errors"
	"math/rand"
	"testing"

	crand "crypto/rand"

	"github.com/avalonbits/shortener/service"
	"github.com/avalonbits/shortener/storage"
)

func TestGenerateShort(t *testing.T) {
	for i := 0; i < 350000; i++ {
		short, err := service.GenerateShort(crand.Reader)
		if err != nil {
			t.Errorf("GenerateShort(): got %v, expected no error", short)
		}
		if len(short) != 8 {
			t.Errorf("short name %q has wrong %d characters, should be 8", short, len(short))
		}
		if _, err := base64.URLEncoding.DecodeString(short); err != nil {
			t.Errorf("base64.URLEncoding.DecodeString(%q): got %v, expected no error", short, err)
		}
	}
}

func TestShortNameFor(t *testing.T) {
	srv := service.NewShortener(
		storage.New(storage.TestDB()),
		crand.Reader,
		1 /*existsRetry=*/)
	ctx := context.Background()
	for i := 0; i < 20000; i++ {
		short, err := srv.ShortNameFor(ctx, "this-is-a-test-long-name")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(short) != 8 {
			t.Errorf("expected short name with 8 characters, got length of %d characters", len(short))
		}
	}
}

type sameReader struct {
}

// Read will either write all 'A's or all 'B's into the buffer.
func (sr *sameReader) Read(buf []byte) (int, error) {
	char := byte('B')
	if rand.Int()%2 == 0 {
		char = byte('A')
	}
	for i := range buf {
		buf[i] = char
	}
	return len(buf), nil
}

func TestShortNameErrors(t *testing.T) {
	srv := service.NewShortener(
		storage.New(storage.TestDB()),
		&sameReader{},
		1000, /*existsRetry=*/
	)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		short, err := srv.ShortNameFor(ctx, "this-is-a-test-long-name")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(short) != 8 {
			t.Errorf("expected short name with 8 characters, got length of %d characters", len(short))
		}
	}

	_, err := srv.ShortNameFor(ctx, "this-will-be-an-error")
	if err == nil {
		t.Errorf("expected an error, got no error.")
	}
	if !errors.Is(err, service.ErrExists) {
		t.Errorf("expected a database error, got %v", err)
	}
}
func TestValidate(t *testing.T) {
	srv := service.NewShortener(
		storage.New(storage.TestDB()),
		crand.Reader,
		1000, /*existsRetry=*/
	)
	ctx := context.Background()

	_, err := srv.ShortNameFor(ctx, "")
	if err == nil {
		t.Errorf("expected error for empty long, got no error")
	}
	if !errors.Is(err, service.ErrEmptyString) {
		t.Errorf("expected empty string error, got %v", err)
	}

	tooLong := randomString(8193)
	_, err = srv.ShortNameFor(ctx, tooLong)
	if !errors.Is(err, service.ErrTooLong) {
		t.Errorf("expoected %v, got %v", service.ErrTooLong, err)
	}

	justRight := tooLong[:len(tooLong)-1]
	_, err = srv.ShortNameFor(ctx, justRight)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func randomString(length int) string {
	b := make([]byte, length)
	const charset = "abcdefghijklmnopqrstuvqwxyz1234567890"
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func TestLongFrom(t *testing.T) {
	shorts := make([]string, 0, 10000)
	value := "this-is-a-test-long-name"
	srv := service.NewShortener(storage.New(storage.TestDB()), crand.Reader, 1 /*existsRetry=*/)
	ctx := context.Background()
	for i := 0; i < cap(shorts); i++ {
		short, err := srv.ShortNameFor(ctx, value)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if len(short) != 8 {
			t.Errorf("expected short name with 8 characters, got length of %d characters", len(short))
		}
		shorts = append(shorts, short)
	}

	rand.Shuffle(len(shorts), func(i, j int) {
		shorts[i], shorts[j] = shorts[j], shorts[i]
	})

	for i := 0; i < len(shorts); i++ {
		long, err := srv.LongFrom(ctx, shorts[i])
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if long != value {
			t.Errorf("%q != %q", long, value)
		}
	}
}

func TestLongFromNotFound(t *testing.T) {
	srv := service.NewShortener(storage.New(storage.TestDB()), crand.Reader, 1 /*existsRetry=*/)
	ctx := context.Background()
	_, err := srv.LongFrom(ctx, "non-existing-short")
	if err == nil {
		t.Errorf("expected an error, got no error.")
	}
	if !errors.Is(err, service.ErrNotFound) {
		t.Errorf("expected not found error, got %v", err)
	}
}

var globalShort string

func BenchmarkGenerateShort(b *testing.B) {
	var localShort string
	var err error

	for i := 0; i < b.N; i++ {
		localShort, err = service.GenerateShort(crand.Reader)
		if err != nil {
			panic(err)
		}
	}
	globalShort = localShort
}

func BenchmarkShortNameFor(b *testing.B) {
	srv := service.NewShortener(storage.New(storage.TestDB()), crand.Reader, 1 /*existsRetry=*/)
	ctx := context.Background()

	var localShort string
	var err error
	for i := 0; i < b.N; i++ {
		localShort, err = srv.ShortNameFor(ctx, "this-is-a-test-long-name")
		if err != nil {
			panic(err)
		}
	}
	globalShort = localShort
}

var globalLong string

func BenchmarkLongFrom(b *testing.B) {
	shorts := make([]string, 0, 10000)
	value := "this-is-a-test-long-name"
	srv := service.NewShortener(storage.New(storage.TestDB()), crand.Reader, 1 /*existsRetry=*/)
	ctx := context.Background()
	for i := 0; i < cap(shorts); i++ {
		short, err := srv.ShortNameFor(ctx, value)
		if err != nil {
			b.Errorf("expected no error, got %v", err)
		}
		if len(short) != 8 {
			b.Errorf("expected short name with 8 characters, got length of %d characters", len(short))
		}
		shorts = append(shorts, short)
	}

	b.Run("InOrder", func(b *testing.B) {
		var localLong string
		var err error
		for i := 0; i < b.N; i++ {
			localLong, err = srv.LongFrom(ctx, shorts[i%len(shorts)])
			if err != nil {
				b.Errorf("expected no error, got %v", err)
			}
		}
		globalLong = localLong
	})

	rand.Shuffle(len(shorts), func(i, j int) {
		shorts[i], shorts[j] = shorts[j], shorts[i]
	})

	b.Run("Random", func(b *testing.B) {
		var localLong string
		var err error
		for i := 0; i < b.N; i++ {
			localLong, err = srv.LongFrom(ctx, shorts[i%len(shorts)])
			if err != nil {
				b.Errorf("expected no error, got %v", err)
			}
		}
		globalLong = localLong
	})
}
