package gin_redisgo_cooldowns

import (
	"github.com/SilverCory/gin-redisgo-cooldowns/redisutils"

	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

func TestIpLimiter(t *testing.T) {
	r := gin.Default()

	redisPool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
		Dial: func() (redis.Conn, error) {
			return redisutils.DialWithDB("tcp", "127.0.0.1:6379", "", "0")
		},
	}

	r.Use(NewRateLimit(redisPool, "ratelimit.IP:", 100, time.Second*5, nil))

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
