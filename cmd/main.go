package main

import (
	"context"
	"fmt"
	"instacart-acc-gen/gen"
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
)

var (
	log     *logrus.Logger
	proxies []string
	codes   []string
)

func init() {
	log = logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "15:04:05",
	})
	proxies, _ = gen.LoadTxtFile("data/proxies.txt")
	codes, _ = gen.LoadTxtFile("data/codes.txt")
}

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config, err := gen.LoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %s", err)
	}

	var count int
	log.Info("enter the number of tasks to run concurrently:")
	if _, err := fmt.Scan(&count); err != nil {
		log.Fatalf("invalid input: %s", err)
	}
	if count >= 120 {
		log.Warn("max concurrent tasks is 120, setting to 120")
		count = 120
	}

	if len(codes) == 0 {
		log.Fatal("no coupon codes found in data/codes.txt")
	}

	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(int64(count))
	for i := 0; i < config.AccountQuantity; i++ {
		wg.Add(1)
		sem.Acquire(ctx, 1)
		go func(i int) {
			defer wg.Done()
			defer sem.Release(1)
			s, err := gen.NewSession(log, ctx, cancel, config, i+1)
			if err != nil {
				log.Fatalf("error creating session: %s", err)
			}
			s.CouponCodes = codes
			s.ProxyList = proxies
			if err := s.GenAccount(); err != nil {
				log.Errorf("error generating account: %s", err)
			}
		}(i)
		time.Sleep(time.Duration(500+rand.Intn(3000)) * time.Millisecond)
	}
	wg.Wait()

	log.Info("tasks completed!")
}
