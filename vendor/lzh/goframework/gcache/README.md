# gcache
gcache 是基于go-cache开源项目修改封装，暴露shareded模式，原理是利用分桶的策略，缓解golang的map共享一把锁的代价，将锁力度下降

### Installation

`go get github.com/wyywawj1991/goframework`

### Usage

#### 单桶Cache模式，共享一个Map
```go
import (
	"fmt"
	"time"

	"github.com/wyywawj1991/goframework/gcache"
)

func main() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	c := gcache.New(5*time.Minute, 10*time.Minute)

	// Set the value of the key "foo" to "bar", with the default expiration time
	c.Set("foo", "bar", cache.DefaultExpiration)

	// Set the value of the key "baz" to 42, with no expiration time
	// (the item won't be removed until it is re-set, or removed using
	// c.Delete("baz")
	c.Set("baz", 42, cache.NoExpiration)

	// Get the string associated with the key "foo" from the cache
	foo, found := c.Get("foo")
	if found {
		fmt.Println(foo)
	}

	// Since Go is statically typed, and cache values can be anything, type
	// assertion is needed when values are being passed to functions that don't
	// take arbitrary types, (i.e. interface{}). The simplest way to do this for
	// values which will only be used once--e.g. for passing to another
	// function--is:
	foo, found := c.Get("foo")
	if found {
		MyFunction(foo.(string))
	}

	// This gets tedious if the value is used several times in the same function.
	// You might do either of the following instead:
	if x, found := c.Get("foo"); found {
		foo := x.(string)
		// ...
	}
	// or
	var foo string
	if x, found := c.Get("foo"); found {
		foo = x.(string)
	}
	// ...
	// foo can then be passed around freely as a string

	// Want performance? Store pointers!
	c.Set("foo", &MyStruct, cache.DefaultExpiration)
	if x, found := c.Get("foo"); found {
		foo := x.(*MyStruct)
			// ...
	}
}
```

#### 分桶Cache模式，每个桶一个Map
```go
import (
	"fmt"
	"time"

	"github.com/wyywawj1991/goframework/gcache"
)
var shardedKeys = []string{
	"f",
	"fo",
	"foo",
	"barf",
	"barfo",
	"foobar",
	"bazbarf",
	"bazbarfo",
	"bazbarfoo",
	"foobarbazq",
	"foobarbazqu",
	"foobarbazquu",
	"foobarbazquux",
}

func main() {
	// Create a cache with a default expiration time of 5 minutes, and which
	// purges expired items every 10 minutes
	tc := NewBucketCache(5*time.Minute, 10*time.Minute, 10)
	for _, v := range shardedKeys {
		tc.Set(v, "value", DefaultExpiration)
	}
}
```

### 单桶Cache缓存支持函数列表

### 分桶Cache缓存支持函数列表

