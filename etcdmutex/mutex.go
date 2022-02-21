package etcdmutex

import (
	"context"
	"errors"
	"go.etcd.io/etcd/api/v3/mvccpb"
)

type EtcdMutex interface {
	Lock(ctx context.Context) (ok bool, err error)
	Unlock(ctx context.Context) (err error)
}

type mutex struct {
	id      string
	lockLey string
	ops     *MutexFactory
}

func (m *mutex) Lock(ctx context.Context) (ok bool, err error) {
	info, has, err := m.ops.get(ctx, m.lockLey)
	if err != nil {
		return
	}
	if has && info.Owner == m.id {
		info.Owner = m.id
		info.Count++
		ok, err =  m.ops.compareAndPut(ctx, m.lockLey, info)
		return
	}
	if !has { // fast path
		trySuccess, putErr := m.ops.compareAndPut(ctx, m.lockLey, lock{Owner: m.id, Count: 1})
		if putErr != nil {
			err = putErr
			return
		}
		if trySuccess {
			ok = trySuccess
			return
		}
	}
	client := m.ops.client
	watchCtx, cancelWatch := context.WithCancel(ctx)
	defer cancelWatch()
	change := client.Watch(watchCtx, m.lockLey)
	for {
		select {
		case resp := <- change:
			for _, v  := range resp.Events {
				if v.Type == mvccpb.DELETE {
					goto TryLock
				}
			}
		case <- ctx.Done():
			return
		}
TryLock:
		ok, err = m.ops.compareAndPut(ctx, m.lockLey, lock{Owner: m.id, Count: 1})
		if err != nil {
			return
		}
		if ok {
			return
		}
	}
}

func (m *mutex) Unlock(ctx context.Context) (err error) {
	info, has, err := m.ops.get(ctx, m.lockLey)
	if err != nil {
		return
	}
	if !has { // lock not exists
		err = errors.New("lock not exists")
		return
	}
	if info.Owner != m.id { // illegal operate
		err = errors.New("illegal operation")
		return
	}
	info.Count--
	if info.Count < 1 {
		_, err = m.ops.compareAndDel(ctx, m.lockLey, info)
		return
	}
	updated, err := m.ops.compareAndPut(ctx, m.lockLey, info)
	if err != nil {
		return
	}
	if !updated { //
		err = errors.New("updated lock failed")
	}
	return
}



