package valueobject

import (
	"crypto/rand"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	monotonicMu sync.Mutex
	entropy     = ulid.Monotonic(rand.Reader, 0)
)

// NewID generates a new ULID string with monotonic entropy.
func NewID() string {
	monotonicMu.Lock()
	defer monotonicMu.Unlock()
	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
}
