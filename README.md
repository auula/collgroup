# collGroup
Wait group for collecting goroutine information.

## 前 言

>在go语言`waitGroup`和`errGroup`都是用来控制`goroutine`的并发的方式，前者只能等待所有`goroutine`执行完成之后再执行`Wait()`函数后面的代码并且不能捕获运行中的错误，而后者能解决在`goroutine`运行出现的错误还能继续，但是只能捕获到第一次出错的`goroutine`的错误信息。有时候我们需要让多个协程在其中几个出错的时候还能正常运行其他的协程，并且还能捕获到出错协程的相关信息，前面2个`waitGroup`和`errGroup`都不能够满足我们的需求，所以打算自己动手实现一个`collectGroup`。


## 需求分析

- 能够支持`context`
- 能够获取错误信息

> 当然我们使用`errGroup`加`channel`也可以实现，但是笔者想自己撸一个单独包。

## 代码实现

Group 是第一个的 collection 结构体，他有3个函数`Go`和`Wait`、`WithContext`

```go

// Group collection group
type Group struct {
	cancel func()
	wg     sync.WaitGroup
	once   sync.Once
	rwm    sync.RWMutex
	Errs   map[string]error
}

```

Go 函数 可以帮你起一个协程运行你的函数，需要你传一个`goroutine`的唯一ID，方便你后面定位到那个任务。

```go
func (g *Group) Go(id string, fn func() error) {
	g.wg.Add(1)
	go func() {
		id := id
		defer g.wg.Done()
		if err := fn(); err != nil {
			// 必须加锁 不然 fatal error: concurrent map writes
			// 写锁
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
}

```
Wait Group 等待函数所有Go函数执行完毕

```go
func (g *Group) Wait() bool {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return len(g.Errs) > 0
}
```

 WithContext 返回一个 Group 和 ctx

```go
func WithContext(ctx context.Context) (*Group, context.Context) {
    // create group parent context & cancel func
    ctx, cancel := context.WithCancel(ctx)
    group := new(Group)
    group.cancel = cancel
    group.Errs = make(map[string]error, 128)
    return group, ctx
}
```

## 应用案例1
我们在执行多个任务的时候，启动了多个协程，但是我们不能确定这些协程在运行的时候会不会出现问题，而出现了什么样的问题，怎么获取到`error`消息，现在我们通过`collectGroup`就可以实现了。

```go
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
		fmt.Println("run all task  successfully!")
	}
}
```

output

```bash
=== RUN   TestCollGroup
    collect_group_test.go:34: task 2 done.
    collect_group_test.go:39: task 3 done.
    collect_group_test.go:29: task 1 done.
    collect_group_test.go:57: Get errors:  map[go-id-3:task 4 running error go-id-4:task 5 running error]
--- PASS: TestCollGroup (4.00s)
PASS
ok      github.com/higker/collgroup     4.012s
```

## 应用案例2

```go

func TestWithContext(t *testing.T) {

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
		// 出错任务
		func() error {
			time.Sleep(3 * time.Second)
			return errors.New("发起用户余额扣款，发生错误")
		},
	}

	for i, t := range tasks {
		group.Go(fmt.Sprintf("go-id-%s", cast.ToString(i)), t)
	}
	// 监听任务出错了一个就返回

	<-ctx.Done()
	t.Log("group exit...任务出，拿到错误消息回滚业务....")
	t.Log("Get errors: ", group.Errs)

}

```

output

```bash
=== RUN   TestWithContext
    collect_group_test.go:77: 更新库存消息....
    collect_group_test.go:82: 发送用户通知.....
    collect_group_test.go:98: group exit...任务出，拿到错误消息回滚业务....
    collect_group_test.go:99: Get errors:  map[go-id-3:发起用户余额扣款，发生错误]
--- PASS: TestWithContext (3.01s)
PASS
ok      github.com/higker/collgroup     3.013s
```
