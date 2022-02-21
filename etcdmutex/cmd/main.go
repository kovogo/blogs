package main

import (
	"context"
	"etcdmutex"
	"github.com/fatih/color"
	etcd "go.etcd.io/etcd/client/v3"
	"math/rand"
	"time"
)

func main() {
	client, err := etcd.New(etcd.Config{
		Endpoints:   []string{""},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	mutexFactory, err := etcdmutex.NewFactory(ctx, client, etcdmutex.WithTTL(10))
	if err != nil {
		panic(err)
	}
	mutex := mutexFactory.Create("Kovogo")
	waitStart := time.Now()
	ok, err := mutex.Lock(ctx)
	if err != nil {
		panic(err)
	}
	if !ok {
		logFail("Lock failed")
		return
	}
	logSuccess("Lock success, after waiting %d ms\n", time.Since(waitStart).Milliseconds())
	time.Sleep(time.Second * time.Duration(rand.Intn(3) + 5))
	rand.Seed(time.Now().Unix())
	if rand.Intn(10) >= 5 {
		panic("Random panic")
	}
	err = mutex.Unlock(ctx)
	if err != nil {
		panic(err)
	}
	logSuccess("Unlock success")
}

func logFail(format string, args...interface{}) {
	color.Yellow(format, args...)
}


func logSuccess(format string, args...interface{}) {
	color.Green(format, args...)
}