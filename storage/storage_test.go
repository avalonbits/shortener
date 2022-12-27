package storage_test

import (
	"context"
	"testing"

	"github.com/avalonbits/shortener/storage"

	_ "github.com/mattn/go-sqlite3"
)

func TestRepo(t *testing.T) {
	createTestCases := []struct {
		name string

		short  string
		long   string
		hasErr bool
	}{
		{name: "create-empty",
			short: "", long: "", hasErr: false},
		{name: "create-valid",
			short: "short", long: "very-long-name", hasErr: false},
		{name: "create-existing",
			short: "short", long: "very-long-name", hasErr: true},
		{name: "create-exisiting-short-different-long",
			short: "short", long: "other-long-name", hasErr: true},
		{name: "create-duplicate-long",
			short: "sh", long: "very-long-name", hasErr: false},
		{name: "create-existing-long-as-short",
			short: "very-long-name", long: "something", hasErr: false},
	}

	queries := storage.New(storage.TestDB())

	for _, tt := range createTestCases {
		err := queries.SetShort(context.Background(), storage.SetShortParams{
			Short: tt.short,
			Longn: tt.long,
		})

		if err == nil && tt.hasErr {
			t.Errorf("Test %q: got no error, expected an error.", tt.name)
		} else if err != nil && !tt.hasErr {
			t.Errorf("Test %q: got %v, expected no error", tt.name, err)
		}
	}

	getTestCases := []struct {
		name string

		short  string
		long   string
		hasErr bool
	}{
		{name: "get-empty",
			short: "", long: "", hasErr: false},
		{name: "get-non-existing",
			short: "non-existing", hasErr: true},
		{name: "get-existing",
			short: "sh", long: "very-long-name", hasErr: false},
		{name: "get-dupllicate",
			short: "short", long: "very-long-name", hasErr: false},
	}

	for _, tt := range getTestCases {
		long, err := queries.GetLong(context.Background(), tt.short)

		if err == nil && tt.hasErr {
			t.Errorf("Test %q: got no error, expected an error.", tt.name)
		} else if err != nil && !tt.hasErr {
			t.Errorf("Test %q: got %v, expected no error", tt.name, err)
		}

		if err == nil && long != tt.long {
			t.Errorf("Test %q: got %q, expected %q", tt.name, long, tt.long)
		}
	}

}
