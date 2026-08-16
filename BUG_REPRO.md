# Bug 复现

## Bug 是什么

私有条目通过检索、活动流和统计泄露。

## 如何触发

运行 `go test ./internal/application -run TestVaultViewsRespectEntryVisibility -count=1`。

## 错误信息

`application_test.go:73: search leaked or hid entries: got 3, want 2`
