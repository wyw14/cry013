# Bug 复现

## Bug 是什么

请求取消后仍产生写入副作用。

## 如何触发

运行 `go test ./internal/application -run TestCanceledCreateBackupLeavesNoSideEffects -count=1`。

## 错误信息

`application_test.go:197: createBackup error = <nil>, want deadline exceeded`
