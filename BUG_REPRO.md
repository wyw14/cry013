# Bug 复现

## Bug 是什么

多步状态变更失败后部分提交。

## 如何触发

运行 `go test ./internal/application -run TestOwnershipTransferFailureKeepsOriginalOwner -count=1`。

## 错误信息

`application_test.go:230: failed transfer changed state`，其中 `vault.OwnerID` 已变为目标成员，但原所有者和目标成员角色仍保留旧值。
