package app

import (
	"crypto/rand"
	"time"

	"github.com/oklog/ulid/v2"
)

func newID(prefix string) string {
	id := ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader)
	return prefix + "_" + id.String()
}
