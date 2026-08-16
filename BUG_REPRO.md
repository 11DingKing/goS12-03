# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

帮我排查一个倒闸操作的安全问题，先不要修改代码。

并网和黑启动都要求值班调度员发起 + 供电站确认，两方确认后才允许执行。现在出现这样的情况：调度员发起一个并网操作，供电站确认，随后调度员在执行前把这个操作取消掉；但取消之后再调用执行接口，操作居然执行成功了——状态从 cancelled 翻成 executed，还写进了执行时间和一条执行审计记录，REST 层返回 200 而不是拒绝。黑启动是一样的现象。

对照下来，没被供电站确认过就取消的操作、以及已经执行过的操作，再调用执行接口都会被正常拒绝，所以这个问题一直没被发现。现场语义上，这等于已经作废的倒闸操作还会被真的执行。

请定位根因：具体是哪个 Go 文件、哪个符号、这个符号的什么错误行为，以及该行为如何一步步导致上面的现象（包括为什么另外两条路径看起来是正常的）。请给出实际代码阅读或定向复现证据。先只提交排查结论，不要改动仓库里的代码。

## 含 Bug 版本

- 仓库：11DingKing/goS12-03
- 仓库地址：https://github.com/11DingKing/goS12-03.git
- parent SHA：5167902e1f47a422b231055012180da6509706aa

## 复现步骤

```bash
git clone -- https://github.com/11DingKing/goS12-03.git bug-repro
cd bug-repro
git checkout --detach 5167902e1f47a422b231055012180da6509706aa
go test ./internal/service ./internal/httpapi -run "^TestCancelledOperationIsNotExecutable$|^TestDualConfirmedOperationStillExecutes$|^TestHTTPCancelledOperationExecuteRejected$" -count=1 -v
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/service ./internal/httpapi -run "^TestCancelledOperationIsNotExecutable$|^TestDualConfirmedOperationStillExecutes$|^TestHTTPCancelledOperationExecuteRejected$" -count=1 -v
=== RUN   TestCancelledOperationIsNotExecutable
=== RUN   TestCancelledOperationIsNotExecutable/grid_connect
    operation_cancel_test.go:45: expected the execute call on cancelled operation OP-1786923210756886292-1 to be rejected, got status executed
=== RUN   TestCancelledOperationIsNotExecutable/black_start
    operation_cancel_test.go:45: expected the execute call on cancelled operation OP-1786923210759499667-1 to be rejected, got status executed
--- FAIL: TestCancelledOperationIsNotExecutable (0.00s)
    --- FAIL: TestCancelledOperationIsNotExecutable/grid_connect (0.00s)
    --- FAIL: TestCancelledOperationIsNotExecutable/black_start (0.00s)
=== RUN   TestDualConfirmedOperationStillExecutes
--- PASS: TestDualConfirmedOperationStillExecutes (0.00s)
FAIL
FAIL	batteryops/internal/service	0.063s
=== RUN   TestHTTPCancelledOperationExecuteRejected
    operation_cancel_http_test.go:58: execute after cancel: expected 409, got 200: {"id":"OP-1786923218588747629-1","type":"grid_connect","dispatcher_id":"dispatcher-1","station_id":"station-1","dispatcher_confirmed":true,"station_confirmed":true,"status":"executed","created_at":"2026-08-16T23:33:38.588789046Z","executed_at":"2026-08-16T23:33:38.592291879Z","audit_trail":[{"timestamp":"2026-08-16T23:33:38.588789046Z","actor":"dispatcher-1","action":"operation_initiated","detail":"grid_connect"},{"timestamp":"2026-08-16T23:33:38.592135463Z","actor":"station-1","action":"station_confirmed","detail":""},{"timestamp":"2026-08-16T23:33:38.592219088Z","actor":"dispatcher-1","action":"operation_cancelled","detail":""},{"timestamp":"2026-08-16T23:33:38.592291879Z","actor":"system","action":"operation_executed","detail":""}]}
--- FAIL: TestHTTPCancelledOperationExecuteRejected (0.01s)
FAIL
FAIL	batteryops/internal/httpapi	0.050s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/service ./internal/httpapi -run "^TestCancelledOperationIsNotExecutable$|^TestDualConfirmedOperationStillExecutes$|^TestHTTPCancelledOperationExecuteRejected$" -count=1 -v
=== RUN   TestCancelledOperationIsNotExecutable
=== RUN   TestCancelledOperationIsNotExecutable/grid_connect
    operation_cancel_test.go:45: expected the execute call on cancelled operation OP-1786923237812779472-1 to be rejected, got status executed
=== RUN   TestCancelledOperationIsNotExecutable/black_start
    operation_cancel_test.go:45: expected the execute call on cancelled operation OP-1786923237812801263-1 to be rejected, got status executed
--- FAIL: TestCancelledOperationIsNotExecutable (0.00s)
    --- FAIL: TestCancelledOperationIsNotExecutable/grid_connect (0.00s)
    --- FAIL: TestCancelledOperationIsNotExecutable/black_start (0.00s)
=== RUN   TestDualConfirmedOperationStillExecutes
--- PASS: TestDualConfirmedOperationStillExecutes (0.00s)
FAIL
FAIL	batteryops/internal/service	0.001s
=== RUN   TestHTTPCancelledOperationExecuteRejected
    operation_cancel_http_test.go:58: execute after cancel: expected 409, got 200: {"id":"OP-1786923238856455666-1","type":"grid_connect","dispatcher_id":"dispatcher-1","station_id":"station-1","dispatcher_confirmed":true,"station_confirmed":true,"status":"executed","created_at":"2026-08-16T23:33:58.856456Z","executed_at":"2026-08-16T23:33:58.856522541Z","audit_trail":[{"timestamp":"2026-08-16T23:33:58.856456Z","actor":"dispatcher-1","action":"operation_initiated","detail":"grid_connect"},{"timestamp":"2026-08-16T23:33:58.856514333Z","actor":"station-1","action":"station_confirmed","detail":""},{"timestamp":"2026-08-16T23:33:58.856519583Z","actor":"dispatcher-1","action":"operation_cancelled","detail":""},{"timestamp":"2026-08-16T23:33:58.856522541Z","actor":"system","action":"operation_executed","detail":""}]}
--- FAIL: TestHTTPCancelledOperationExecuteRejected (0.00s)
FAIL
FAIL	batteryops/internal/httpapi	0.001s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

准确定位 internal/domain/operation.go 中的 (*Operation).CanExecute（经 (*Operation).Execute 与 service.(*Service).ExecuteOperation 暴露）
说明执行门禁只排除“已执行”状态、未要求仍处于 confirmed 状态，且 Cancel 不回退双方确认标记，从而放行已取消操作的完整因果链
解释未经供电站确认即取消、以及重复执行两条路径为何仍被正确拒绝
结论附实际代码阅读或定向复现证据；目标仓库保持零改动
