// Copyright 2021 The higker Authors. All rights reserved.
// license that can be found in the LICENSE file.
// reference https://github.com/golang/sync/blob/09787c993a3a/errgroup/errgroup.go

// Package collgroup provides synchronization, error propagation, and Context
// cancelation for groups of goroutines working on subtasks of a common task
// collecting goroutine task information
package collgroup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime/pprof"
	"testing"
	"time"

	"github.com/spf13/cast"
)

type task func() error

func TestCollGroup(t *testing.T) {

	// 创建一个collectGroup
	g := new(Group)
	// 模拟多任务
	tasks := []task{
		func() error {
			time.Sleep(4 * time.Second)
			t.Log("task 1 done.")
			return nil
		},
		func() error {
			time.Sleep(2 * time.Second)
			t.Log("task 2 done.")
			return nil
		},
		func() error {
			time.Sleep(3 * time.Second)
			t.Log("task 3 done.")
			return nil
		},
		// 出错任务
		func() error {
			time.Sleep(3 * time.Second)
			return errors.New("task 4 running error")
		},
		func() error {
			time.Sleep(3 * time.Second)
			return errors.New("task 5 running error")
		},
	}
	g.Errs = make(map[string]error, cap(tasks))
	for i, t := range tasks {
		g.Go(fmt.Sprintf("go-id-%s", cast.ToString(i)), t)
	}
	if g.Wait() {
		t.Log("Get errors: ", g.Errs)
	} else {
		t.Log("run all task  successfully!")
	}
}

func TestWithContext(t *testing.T) {
	if true {
		file, err := os.Create("./cpu.pprof")
		if err != nil {
			fmt.Printf("create cpu pprof failed, err:%v\n", err)
			return
		}
		pprof.StartCPUProfile(file)
		defer pprof.StopCPUProfile()
	}
	for i := 0; i < 8; i++ {
		go Test(t)
	}
	time.Sleep(5 * time.Second)
	if true {
		file, err := os.Create("./mem.pprof")
		if err != nil {
			fmt.Printf("create mem pprof failed, err:%v\n", err)
			return
		}
		pprof.WriteHeapProfile(file)
		file.Close()
	}

}

func Test(t *testing.T) {
	// 创建一个errGroup
	group, ctx := WithContext(context.Background())
	// 模拟多任务
	tasks := []task{
		func() error {
			time.Sleep(4 * time.Second)
			t.Log("向订单表加入消息....")
			return nil
		},
		func() error {
			time.Sleep(2 * time.Second)
			t.Log("更新库存消息....")
			return nil
		},
		func() error {
			time.Sleep(3 * time.Second)
			t.Log("发送用户通知.....")
			return nil
		},
	}

	for i, t := range tasks {
		group.Go(fmt.Sprintf("go-id-%s", cast.ToString(i)), t)
	}
	// group.Wait()
	// 监听任务出错了一个就返回
	<-ctx.Done()
	if len(group.Errs) > 0 {
		t.Log("group exit...任务出错，拿到错误消息回滚业务....")
		t.Log("Get errors: ", group.Errs)
	}
}
