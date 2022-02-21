package etcdmutex

import (
	"context"
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	etcd "go.etcd.io/etcd/client/v3"
)

const (
	defaultTTL = 8
)

type MutexFactory struct {
	*FactoryOption
	ctx    context.Context
	client *etcd.Client
	lease  etcd.LeaseID
}

type FactoryOption struct {
	ttl   int64
	idGen func() string
}

type Opt func(option *FactoryOption)

type lock struct {
	Owner   string `json:"owner,omitempty"`
	Count   int    `json:"count,omitempty"`
	Version int64  `json:"-"`
}

func WithTTL(ttl int64) Opt {
	return func(option *FactoryOption) {
		option.ttl = ttl
	}
}

func WithIdGenerator(fn func() string) Opt {
	return func(option *FactoryOption) {
		option.idGen = fn
	}
}

func generateUUID() string {
	return fmt.Sprintf("etcdmutex-%s", uuid.NewV4())
}

var defaultOption = FactoryOption{
	ttl:   defaultTTL,
	idGen: generateUUID,
}

func NewFactory(ctx context.Context, client *etcd.Client, opts ...Opt) (factory *MutexFactory, err error) {
	option := defaultOption
	for _, v := range opts {
		v(&option)
	}
	factory = &MutexFactory{
		ctx:           ctx,
		client:        client,
		FactoryOption: &option,
	}
	if option.ttl > 0 {
		lease, grantErr := client.Grant(ctx, option.ttl)
		if grantErr != nil {
			err = grantErr
			return
		}
		factory.lease = lease.ID
		err = factory.keepAlive(ctx, factory.lease)
		if err != nil {
			return
		}
	}
	return
}

func (m *MutexFactory) keepAlive(ctx context.Context, leaseId etcd.LeaseID) error {
	respCh, err := m.client.KeepAlive(ctx, leaseId)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case _ = <-respCh:
				continue
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func (m *MutexFactory) get(ctx context.Context, key string) (info lock, has bool, err error) {
	resp, err := m.client.Get(ctx, key)
	if err != nil {
		return
	}
	if len(resp.Kvs) == 0 {
		return
	}
	kv := resp.Kvs[0]
	err = json.Unmarshal(kv.Value, &info)
	if err != nil {
		return
	}
	info.Version = kv.ModRevision
	has = true
	return
}

func (m *MutexFactory) compareAndPut(ctx context.Context, key string, info lock) (ok bool, err error) {
	bytes, err := json.Marshal(&info)
	if err != nil {
		return
	}
    resp, err := m.client.Txn(ctx).
		If(etcd.Compare(etcd.ModRevision(key),  "=", info.Version)).
		Then(etcd.OpPut(key, string(bytes), etcd.WithLease(m.lease))).
		Commit()
	if err != nil {
		return
	}
	ok = resp.Succeeded
	return
}

func (m *MutexFactory) compareAndDel(ctx context.Context, key string, info lock) (ok bool, err error) {
	resp, err := m.client.Txn(ctx).
		If(etcd.Compare(etcd.ModRevision(key), "=", info.Version)).
		Then(etcd.OpDelete(key)).Commit()
	if err != nil {
		return
	}
	ok = resp.Succeeded
	return
}

func (m *MutexFactory) Create(key string) EtcdMutex {
	return &mutex{
		id:      m.idGen(),
		lockLey: key,
		ops:     m,
	}
}

func (m *MutexFactory) Close() (err error) {
	return m.client.Close()
}


