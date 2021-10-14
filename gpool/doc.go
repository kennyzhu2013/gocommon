package gpool

// 整个gpool实现思路参考：https://strikefreedom.top/high-performance-implementation-of-goroutine-pool
// Go 调度器工作时会维护两种用来保存 G 的任务队列：一种是一个 Global 任务队列，一种是每个 P 维护的 Local 任务队列。
//
//当通过 go 关键字创建一个新的 goroutine 的时候，它会优先被放入 P 的本地队列。为了运行 goroutine，M 需要持有（绑定）一个 P，接着 M 会启动一个 OS 线程，循环从 P 的本地队列里取出一个 goroutine 并执行。当然还有上文提及的 work-stealing 调度算法：
// 当 M 执行完了当前 P 的 Local 队列里的所有 G 后，P 也不会就这么在那躺尸啥都不干，它会先尝试从 Global 队列寻找 G 来执行，如果 Global 队列为空，它会随机挑选另外一个 P，从它的队列里中拿走一半的 G 到自己的队列中执行。
//
//如果一切正常，调度器会以上述的那种方式顺畅地运行，但这个世界没这么美好，总有意外发生，以下分析 goroutine 在两种例外情况下的行为。
//
//Go runtime 会在下面的 goroutine 被阻塞的情况下运行另外一个 goroutine：
//
// blocking syscall (for example opening a file)
// network input
// channel operations
// primitives in the sync package
