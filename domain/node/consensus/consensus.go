package consensus // 或一个共享的接口定义包

// LogEntry 结构已在 3.2.3 节定义，包含 Term, Index, Command
// type LogEntry struct {
//     Term    int32
//     Index   int64
//     Command []byte // 实际的状态机命令
// }

// FSM 是上层状态机需要实现的接口
type FSM interface {
    // Apply 方法负责将一条已经由 Raft 集群确认提交的日志条目应用到状态机中。
    // 此方法的实现必须是完全确定性的：对于任何给定的初始状态和相同的日志条目序列，
    // 它必须始终产生完全相同的最终状态和相同的输出（如果方法有输出的话）。
    // Raft 模块会在日志条目被安全提交 (commitIndex 更新) 后，严格按照日志索引 (Index)
    // 的顺序，串行调用此方法。
    // Apply 方法的返回值可以用于向客户端提供操作结果的反馈，或者如果不需要反馈，则可以为 nil。
    // 返回的 error 表示状态机在应用该条目时发生内部错误，这通常是严重问题。
    Apply(entry LogEntry) (applicationResult interface{}, err error)

    // --- 以下为可选的快照 (Snapshot) 相关接口，用于实现日志压缩功能 ---

    // GetSnapshot 方法用于获取当前状态机的快照。
    // Raft 模块可能会定期调用此方法（或由其他策略触发），以获取状态机当前状态的
    // 紧凑表示（字节流）。这个快照随后可以被持久化存储，或在 InstallSnapshot RPC
    // 中发送给滞后的 Follower 节点，以帮助它们快速追赶进度，避免传输大量日志条目。
    // 返回的 []byte 是快照数据，error 表示获取快照时发生错误。
    // GetSnapshot() (snapshotData []byte, err error)

    // ApplySnapshot 方法用于从给定的快照字节流中恢复状态机的状态。
    // 当一个 Follower 节点通过 InstallSnapshot RPC 接收到来自 Leader 的快照数据后，
    // Raft 模块会调用此方法，将快照数据传递给 FSM，FSM 负责解析这些数据并
    // 重置其内部状态至快照所代表的状态点。
    // 此操作完成后，FSM 的状态应与 Leader 在生成快照时的状态一致。
    // ApplySnapshot(snapshotData []byte) error
}