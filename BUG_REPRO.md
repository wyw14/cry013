# Bug 复现

## Bug 是什么

条目并发幂等创建失效。

## 如何触发

运行 `go test -race ./internal/application -run TestRepeatedCreateReturnsSingleEntry -count=1`。

## 错误信息

`application_test.go:175: idempotent requests returned different IDs: <id-1> and <id-2>`
