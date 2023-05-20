package gcache

import (
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io"
	"math"
	"math/big"
	insecurerand "math/rand"
	"os"
	"time"
)

const (
	// For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
)

// This is an experimental and unexported (for now) attempt at making a cache
// with better algorithmic complexity than the standard one, namely by
// preventing write locks of the entire cache when an item is added. As of the
// time of writing, the overhead of selecting buckets results in cache
// operations being about twice as slow as for the standard cache with small
// total cache sizes, and faster for larger ones.
//
// See cache_test.go for a few benchmarks.

type shardedCache struct {
	seed    uint32
	m       uint32
	cs      []*cache
	janitor *shardedJanitor
}

// djb2 with better shuffling. 5x faster than FNV with the hash.Hash overhead.
func djb33(seed uint32, k string) uint32 {
	var (
		l = uint32(len(k))
		d = 5381 + seed + l
		i = uint32(0)
	)
	// Why is all this 5x faster than a for loop?
	if l >= 4 {
		for i < l-4 {
			d = (d * 33) ^ uint32(k[i])
			d = (d * 33) ^ uint32(k[i+1])
			d = (d * 33) ^ uint32(k[i+2])
			d = (d * 33) ^ uint32(k[i+3])
			i += 4
		}
	}
	switch l - i {
	case 1:
	case 2:
		d = (d * 33) ^ uint32(k[i])
	case 3:
		d = (d * 33) ^ uint32(k[i])
		d = (d * 33) ^ uint32(k[i+1])
	case 4:
		d = (d * 33) ^ uint32(k[i])
		d = (d * 33) ^ uint32(k[i+1])
		d = (d * 33) ^ uint32(k[i+2])
	}
	return d ^ (d >> 16)
}

func (sc *shardedCache) bucket(k string) *cache {
	if sc.m == 1 {
		return sc.cs[0]
	}
	return sc.cs[djb33(sc.seed, k)%sc.m]
}
func (sc *shardedCache) SetRecover(k string, x interface{}, e int64) {
	sc.bucket(k).setRecover(k, x, e)
}

func (sc *shardedCache) Set(k string, x interface{}, d time.Duration) {
	sc.bucket(k).set(k, x, d)
}

func (sc *shardedCache) SetDefault(k string, x interface{}) {
	sc.bucket(k).set(k, x, DefaultExpiration)
}

func (sc *shardedCache) Add(k string, x interface{}, d time.Duration) error {
	return sc.bucket(k).add(k, x, d)
}

func (sc *shardedCache) Replace(k string, x interface{}, d time.Duration) error {
	return sc.bucket(k).replace(k, x, d)
}

func (sc *shardedCache) Get(k string) (interface{}, bool) {
	return sc.bucket(k).get(k)
}

func (sc *shardedCache) GetWithExpiration(k string) (interface{}, time.Time, bool) {
	return sc.bucket(k).getWithExpiration(k)
}

func (sc *shardedCache) Increment(k string, n int64) error {
	return sc.bucket(k).increment(k, n)
}

func (sc *shardedCache) IncrementInt(k string, n int) (int, error) {
	return sc.bucket(k).incrementInt(k, n)
}

func (sc *shardedCache) IncrementInt8(k string, n int8) (int8, error) {
	return sc.bucket(k).incrementInt8(k, n)
}

func (sc *shardedCache) IncrementInt16(k string, n int16) (int16, error) {
	return sc.bucket(k).incrementInt16(k, n)
}

func (sc *shardedCache) IncrementInt32(k string, n int32) (int32, error) {
	return sc.bucket(k).incrementInt32(k, n)
}

func (sc *shardedCache) IncrementInt64(k string, n int64) (int64, error) {
	return sc.bucket(k).incrementInt64(k, n)
}

func (sc *shardedCache) IncrementUint(k string, n uint) (uint, error) {
	return sc.bucket(k).incrementUint(k, n)
}

func (sc *shardedCache) IncrementUintptr(k string, n uintptr) (uintptr, error) {
	return sc.bucket(k).incrementUintptr(k, n)
}

func (sc *shardedCache) IncrementUint8(k string, n uint8) (uint8, error) {
	return sc.bucket(k).incrementUint8(k, n)
}

func (sc *shardedCache) IncrementUint16(k string, n uint16) (uint16, error) {
	return sc.bucket(k).incrementUint16(k, n)
}

func (sc *shardedCache) IncrementUint32(k string, n uint32) (uint32, error) {
	return sc.bucket(k).incrementUint32(k, n)
}

func (sc *shardedCache) IncrementUint64(k string, n uint64) (uint64, error) {
	return sc.bucket(k).incrementUint64(k, n)
}

func (sc *shardedCache) IncrementFloat(k string, n float64) error {
	return sc.bucket(k).incrementFloat(k, n)
}

func (sc *shardedCache) IncrementFloat32(k string, n float32) (float32, error) {
	return sc.bucket(k).incrementFloat32(k, n)
}

func (sc *shardedCache) IncrementFloat64(k string, n float64) (float64, error) {
	return sc.bucket(k).incrementFloat64(k, n)
}

func (sc *shardedCache) Decrement(k string, n int64) error {
	return sc.bucket(k).decrement(k, n)
}

func (sc *shardedCache) DecrementInt(k string, n int) (int, error) {
	return sc.bucket(k).decrementInt(k, n)
}

func (sc *shardedCache) DecrementInt8(k string, n int8) (int8, error) {
	return sc.bucket(k).decrementInt8(k, n)
}

func (sc *shardedCache) DecrementInt16(k string, n int16) (int16, error) {
	return sc.bucket(k).decrementInt16(k, n)
}

func (sc *shardedCache) DecrementInt32(k string, n int32) (int32, error) {
	return sc.bucket(k).decrementInt32(k, n)
}

func (sc *shardedCache) DecrementInt64(k string, n int64) (int64, error) {
	return sc.bucket(k).decrementInt64(k, n)
}

func (sc *shardedCache) DecrementUint(k string, n uint) (uint, error) {
	return sc.bucket(k).decrementUint(k, n)
}

func (sc *shardedCache) DecrementUintptr(k string, n uintptr) (uintptr, error) {
	return sc.bucket(k).decrementUintptr(k, n)
}

func (sc *shardedCache) DecrementUint8(k string, n uint8) (uint8, error) {
	return sc.bucket(k).decrementUint8(k, n)
}

func (sc *shardedCache) DecrementUint16(k string, n uint16) (uint16, error) {
	return sc.bucket(k).decrementUint16(k, n)
}

func (sc *shardedCache) DecrementUint32(k string, n uint32) (uint32, error) {
	return sc.bucket(k).decrementUint32(k, n)
}

func (sc *shardedCache) DecrementUint64(k string, n uint64) (uint64, error) {
	return sc.bucket(k).decrementUint64(k, n)
}

func (sc *shardedCache) DecrementFloat(k string, n float64) error {
	return sc.bucket(k).decrementFloat(k, n)
}

func (sc *shardedCache) DecrementFloat32(k string, n float32) (float32, error) {
	return sc.bucket(k).decrementFloat32(k, n)
}

func (sc *shardedCache) DecrementFloat64(k string, n float64) (float64, error) {
	return sc.bucket(k).decrementFloat64(k, n)
}

func (sc *shardedCache) Delete(k string) {
	sc.bucket(k).delete(k)
}

func (sc *shardedCache) DeleteExpired() {
	for _, v := range sc.cs {
		v.deleteExpired()
	}
}

func (sc *shardedCache) OnEvicted(f func(string, interface{})) {
	for _, v := range sc.cs {
		v.addEvicted(f)
	}
}

// Returns the items in the cache. This may include items that have expired,
// but have not yet been cleaned up. If this is significant, the Expiration
// fields of the items should be checked. Note that explicit synchronization
// is needed to use a cache and its corresponding Items() return values at
// the same time, as the maps are shared.
func (sc *shardedCache) Items() []map[string]Item {
	res := make([]map[string]Item, len(sc.cs))
	for i, v := range sc.cs {
		res[i] = v.getItems()
	}
	return res
}

func (sc *shardedCache) ItemsCount() []int {
	res := make([]int, len(sc.cs))
	for i, v := range sc.cs {
		res[i] = v.getItemCount()
	}
	return res
}

func (sc *shardedCache) Flush() {
	for _, v := range sc.cs {
		v.flush()
	}
}

func (sc *shardedCache) SaveFile(fname string) error {
	for index, v := range sc.cs {
		v.saveFile(fmt.Sprintf("%s_%d", fname, index))
	}
	return nil
}

func (sc *shardedCache) LoadFile(fname string) error {
	fp, err := os.Open(fname)
	if err != nil {
		return err
	}
	err = sc.Load(fp)
	if err != nil {
		fp.Close()
		return err
	}
	return fp.Close()
}

func (sc *shardedCache) Load(r io.Reader) error {
	dec := gob.NewDecoder(r)
	items := map[string]Item{}
	err := dec.Decode(&items)
	if err == nil {
		for k, v := range items {
			if !v.Expired() {
				sc.SetRecover(k, v.Object, v.Expiration)
			}
		}
	}
	return err
}

type shardedJanitor struct {
	Interval time.Duration
	stop     chan bool
}

func (j *shardedJanitor) Run(sc *shardedCache) {
	j.stop = make(chan bool)
	tick := time.Tick(j.Interval)
	for {
		select {
		case <-tick:
			sc.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}

func stopShardedJanitor(sc *BucketCache) {
	sc.janitor.stop <- true
}

func runShardedJanitor(sc *shardedCache, ci time.Duration) {
	j := &shardedJanitor{
		Interval: ci,
	}
	sc.janitor = j
	go j.Run(sc)
}

func newShardedCache(n int, de time.Duration) *shardedCache {
	max := big.NewInt(0).SetUint64(uint64(math.MaxUint32))
	rnd, err := rand.Int(rand.Reader, max)
	var seed uint32
	if err != nil {
		os.Stderr.Write([]byte("WARNING: go-cache's newShardedCache failed to read from the system CSPRNG (/dev/urandom or equivalent.) Your system's security may be compromised. Continuing with an insecure seed.\n"))
		seed = insecurerand.Uint32()
	} else {
		seed = uint32(rnd.Uint64())
	}
	sc := &shardedCache{
		seed: seed,
		m:    uint32(n),
		cs:   make([]*cache, n),
	}
	for i := 0; i < n; i++ {
		c := &cache{
			defaultExpiration: de,
			items:             map[string]Item{},
		}
		sc.cs[i] = c
	}
	return sc
}
