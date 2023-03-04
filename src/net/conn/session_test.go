package conn

import (
	"context"
	"math/rand"
	"testing"
	"time"
)

func randInt(a, b int) int {
	return rand.Intn(b-a+1) + a
}

func TestSession_CreateSession(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	m := NewSessionManager(context.Background())
	const (
		maxCount = 100
	)

	id := uint64(0)

	for i := 0; i < maxCount; i++ {
		id++

		if (id % 5) == 0 {
			//sid := uint64(randInt(0, int(id)))
			//s := m.GetSessionById(sid)
			//go m.RemoveSession(s)
		}

		go m.NewSession(context.Background(), id, nil)
	}
}
