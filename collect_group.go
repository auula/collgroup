// Copyright 2021 The higker Authors. All rights reserved.
// license that can be found in the LICENSE file.
// reference https://github.com/golang/sync/blob/09787c993a3a/errgroup/errgroup.go

// Package collgroup provides synchronization, error propagation, and Context
// cancelation for groups of goroutines working on subtasks of a common task
// collecting goroutine task information
package collgroup

import (
	"context"
	"sync"
)

// Group collection group
type Group struct {
	cancel func()
	wg     sync.WaitGroup
	once   sync.Once
	rwm    sync.RWMutex
	Errs   map[string]error
}

// WithContext 返回一个 Group 和 ctx
func WithContext(ctx context.Context) (*Group, context.Context) {
	// create group parent context & cancel func
	ctx, cancel := context.WithCancel(ctx)
	group := new(Group)
	group.cancel = cancel
	group.Errs = make(map[string]error, 128)
	return group, ctx
}

// Go 函数 可以帮你起一个协程运行你的函数
func (g *Group) Go(id string, fn func() error) {
	g.wg.Add(1)
	go func() {
		//id := id
		defer g.wg.Done()
		if err := fn(); err != nil {
			// 写锁必须加锁 不然 fatal error: concurrent map writes
			g.rwm.Lock()
			g.Errs[id] = err
			g.rwm.Unlock()
			g.once.Do(func() {
				// 只执行一次
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
	// heartBeat 这里是指的是 group有协程在运行
	// 并且没有发送错误的时候在
	// 会发生无法退出的情况 使用通过heartBeat来解决
	go g.Wait()
}

// Wait Group 等待函数
func (g *Group) Wait() bool {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return len(g.Errs) > 0
}
