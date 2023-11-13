package op

import (
	"time"

	"github.com/bluele/gcache"
	"github.com/zijiren233/gencontainer/synccache"
)

func Init(size int) error {
	roomCache = synccache.NewSyncCache[string, *Room](time.Minute*5, synccache.WithDeletedCallback[string, *Room](func(v *Room) {
		v.close()
	}))
	userCache = gcache.New(size).
		LRU().
		Build()

	return nil
}
