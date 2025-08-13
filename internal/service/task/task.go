package task

import (
	"container/heap"
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

// DelayedTask 延时任务定义
type DelayedTask struct {
	ID         string
	ExecuteAt  time.Time // 执行时间点
	TaskFunc   func(ctx context.Context) error
	Name       string
	DataType   string
	MailType   int
	GenerateAt time.Time //生成时间
	LedgerName string
}

// taskHeap 任务堆实现
type taskHeap []*DelayedTask

func (h taskHeap) Len() int            { return len(h) }
func (h taskHeap) Less(i, j int) bool  { return h[i].ExecuteAt.Before(h[j].ExecuteAt) }
func (h taskHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *taskHeap) Push(x interface{}) { *h = append(*h, x.(*DelayedTask)) }
func (h *taskHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// DelayedTaskScheduler 优化的延时任务调度器
type DelayedTaskScheduler struct {
	taskQueue chan *DelayedTask // 任务输入队列
	taskHeap  *taskHeap         // 优先级队列管理待执行任务
	readyChan chan *DelayedTask // 准备执行的任务通道
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	running   atomic.Bool
	taskMap   sync.Map // 任务ID到任务的映射
}

// NewDelayedTaskScheduler 创建优化后的调度器
func NewDelayedTaskScheduler() *DelayedTaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())
	th := &taskHeap{}
	heap.Init(th)

	return &DelayedTaskScheduler{
		taskQueue: make(chan *DelayedTask, 1000),
		taskHeap:  th,
		readyChan: make(chan *DelayedTask, 100),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start 启动调度器
func (s *DelayedTaskScheduler) Start(workers int) {
	if !s.running.CompareAndSwap(false, true) {
		return
	}

	// 启动任务分发器
	s.wg.Add(1)
	go s.taskDispatcher()

	// 启动工作池
	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go s.worker(i)
	}
}

// taskDispatcher 任务分发器 - 优化使用时间堆
func (s *DelayedTaskScheduler) taskDispatcher() {
	defer s.wg.Done()
	defer log.Println("分发器已停止")

	// 初始定时器
	timer := time.NewTimer(time.Hour)
	defer timer.Stop()

	// 重置定时器到最近任务
	resetTimer := func() {
		if s.taskHeap.Len() == 0 {
			timer.Reset(time.Hour)
			return
		}

		task := (*s.taskHeap)[0]
		duration := time.Until(task.ExecuteAt)
		if duration < 0 {
			duration = 0
		}
		timer.Reset(duration)
	}

	// 初始设置超长时间
	if !timer.Stop() {
		<-timer.C
	}
	resetTimer()

	for {
		select {
		case <-s.ctx.Done():
			return
		case newTask := <-s.taskQueue:
			// 添加新任务到堆
			heap.Push(s.taskHeap, newTask)
			s.taskMap.Store(newTask.ID, newTask)

			// 如果新任务是最近的任务，重置定时器
			if s.taskHeap.Len() == 1 || newTask.ExecuteAt.Before((*s.taskHeap)[0].ExecuteAt) {
				resetTimer()
			}

		case <-timer.C:
			now := time.Now()

			// 处理所有到期任务
			for s.taskHeap.Len() > 0 {
				task := (*s.taskHeap)[0]
				if task.ExecuteAt.After(now) {
					break
				}

				// 分发到工作通道
				select {
				case s.readyChan <- task:
					heap.Pop(s.taskHeap)
					s.taskMap.Delete(task.ID)
				case <-s.ctx.Done():
					return
				default:
					// 通道满时等待一小段时间
					time.Sleep(10 * time.Millisecond)
				}
			}

			// 重置定时器
			resetTimer()
		}
	}
}

// worker 工作协程 - 添加上下文超时控制
func (s *DelayedTaskScheduler) worker(id int) {
	defer s.wg.Done()
	defer log.Printf("工作协程 %d 已停止", id)

	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.readyChan:
			// 执行任务，限制最大执行时间
			taskCtx, cancel := context.WithTimeout(s.ctx, 2*time.Minute)

			// 执行任务
			start := time.Now()
			err := task.TaskFunc(taskCtx)
			duration := time.Since(start)

			// 取消上下文（无论任务是否完成）
			cancel()

			if err != nil {
				log.Printf("任务 %s 执行失败 (耗时 %v): %v", task.ID, duration, err)
			} else {
				log.Printf("工作协程 %d 成功执行任务 %s (耗时 %v)", id, task.ID, duration)
			}
		}
	}
}

// Schedule 添加延时任务 (支持秒级精度)
func (s *DelayedTaskScheduler) Schedule(id string, delay time.Duration, task func(ctx context.Context) error) {
	if delay < 0 {
		delay = 0
	}

	select {
	case s.taskQueue <- &DelayedTask{
		ID:        id,
		ExecuteAt: time.Now().Add(delay),
		TaskFunc:  task,
	}:
	case <-time.After(100 * time.Millisecond):
		log.Printf("任务 %s 添加超时", id)
	}
}

func (s *DelayedTaskScheduler) Stop() {
	if !s.running.CompareAndSwap(true, false) {
		return
	}

	s.cancel()

	// 快速清空任务队列
	go func() {
		for range s.taskQueue {
			// 只是清空，不处理
		}
	}()

	// 等待协程结束
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("调度器已停止")
	case <-time.After(10 * time.Second):
		log.Println("警告: 停止超时，强制退出")
	}
}
