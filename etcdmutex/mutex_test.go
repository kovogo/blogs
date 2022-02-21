package etcdmutex

import (
	"context"
	etcd "go.etcd.io/etcd/client/v3"
	"testing"
	"time"
)

func Benchmark_Mutex(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		client, err := etcd.New(etcd.Config{
			Endpoints:   []string{""},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			panic(err)
		}
		ctx := context.Background()
		mutexFactory, err :=NewFactory(ctx, client, WithTTL(10))
		if err != nil {
			panic(err)
		}
		mux := mutexFactory.Create("KeSanGo")
		for pb.Next() {
			ok, err := mux.Lock(ctx)
			if err != nil {
				panic(err)
			}
			if !ok {
				b.Errorf("lock failed")
				return
			}
			err = mux.Unlock(ctx)
			if err != nil {
				panic(err)
			}
		}
	})
}