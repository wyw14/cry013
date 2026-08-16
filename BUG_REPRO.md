# Bug 复现

## Bug 是什么

并发令牌重放。

## 如何触发

运行 `go test -race ./internal/application -run TestRefreshOnlyOneConcurrentRotationSucceeds -count=1`。

## 错误信息

`application_test.go:136: refresh token rotated 24 times, want exactly 1`
