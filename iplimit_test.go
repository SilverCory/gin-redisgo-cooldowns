package gin_redisgo_cooldowns

import (
	"github.com/go-redis/redis/v8"

	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestIpLimiter(t *testing.T) {
	r := gin.Default()

	rdb := redis.NewClient(&redis.Options{
		Addr:    "127.0.0.1:6379",
		DB:       0,
	})

	r.Use(NewRateLimit(rdb, "ratelimit.IP:", 100, time.Second*5, nil))

	r.GET("/", func(c *gin.Context) {
		c.String(200, "OK")
	})

	go r.Run(":9999")

	for i := 0; i < 102; i++ {
		c := &http.Client{}

		resp, e := c.Get("http://127.0.0.1:9999")
		if e != nil {
			t.Error("Error during requests ", e.Error())
			return
		}

		switch {
		case i < 100:
			break
		case i == 100:
			if resp.StatusCode != 429 {
				t.Error("Threashold break not detected")
			} else {
				time.Sleep(time.Second * 5)
			}
			break
		case i == 101:
			if resp.StatusCode != 200 {
				t.Error("Unnecessary block")
			}
			break
		}
	}
}
