package storage

import (
	_ "embed"
)

//go:embed schema.sql
var Schema string
