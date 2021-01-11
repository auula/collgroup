// Copyright 2021 The higker Authors. All rights reserved.
// license that can be found in the LICENSE file.
// reference https://github.com/golang/sync/blob/09787c993a3a/errgroup/errgroup.go

// Package collgroup provides synchronization, error propagation, and Context
// cancelation for groups of goroutines working on subtasks of a common task
// collecting goroutine task information
package collgroup

import (
	"errors"
	"fmt"
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
			fmt.Println("task 1 done.")
			return nil
		},
		func() error {
			time.Sleep(2 * time.Second)
			fmt.Println("task 2 done.")
			return nil
		},
		func() error {
			time.Sleep(3 * time.Second)
			fmt.Println("task 3 done.")
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
		fmt.Println("Get errors: ", g.Errs)
	} else {
		fmt.Println("Get all num successfully!")
	}
}
