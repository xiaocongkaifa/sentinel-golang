package base

import (
	"fmt"
	"sync"
)

type BlockType int32

const (
	BlockTypeUnknown BlockType = iota
	BlockTypeFlow
	BlockTypeCircuitBreaking
	BlockTypeSystemFlow
)

var resultPool = sync.Pool{
	New: func() interface{} {
		return NewTokenResultEmpty()
	},
}

func (t BlockType) String() string {
	switch t {
	case BlockTypeUnknown:
		return "Unknown"
	case BlockTypeFlow:
		return "FlowControl"
	case BlockTypeCircuitBreaking:
		return "CircuitBreaking"
	case BlockTypeSystemFlow:
		return "System"
	default:
		return fmt.Sprintf("%d", t)
	}
}

type TokenResultStatus int32

const (
	ResultStatusPass TokenResultStatus = iota
	ResultStatusBlocked
	ResultStatusShouldWait
)

type TokenResult struct {
	status TokenResultStatus

	blockErr *BlockError
	waitMs   uint64
}

func (r *TokenResult) IsPass() bool {
	return r.status == ResultStatusPass
}

func (r *TokenResult) IsBlocked() bool {
	return r.status == ResultStatusBlocked
}

func (r *TokenResult) Status() TokenResultStatus {
	return r.status
}

func (r *TokenResult) BlockError() *BlockError {
	return r.blockErr
}

func (r *TokenResult) WaitMs() uint64 {
	return r.waitMs
}

func (r *TokenResult) String() string {
	var blockMsg string
	if r.blockErr == nil {
		blockMsg = "none"
	} else {
		blockMsg = r.blockErr.Error()
	}
	return fmt.Sprintf("TokenResult{status=%d, blockErr=%s, waitMs=%d}", r.status, blockMsg, r.waitMs)
}

func NewTokenResultEmpty() *TokenResult {
	return &TokenResult{
		status:   ResultStatusPass,
		blockErr: nil,
		waitMs:   0,
	}
}

func NewTokenResultPass() *TokenResult {
	return resultPool.Get().(*TokenResult)
}

func NewTokenResultBlocked(blockType BlockType, blockMsg string) *TokenResult {
	result := resultPool.Get().(*TokenResult)
	result.status = ResultStatusBlocked
	result.blockErr = NewBlockError(blockType, blockMsg)
	return result
}

func NewTokenResultBlockedWithCause(blockType BlockType, blockMsg string, rule SentinelRule, snapshot interface{}) *TokenResult {
	result := resultPool.Get().(*TokenResult)
	result.status = ResultStatusBlocked
	result.blockErr = NewBlockErrorWithCause(blockType, blockMsg, rule, snapshot)
	return result
}

func NewTokenResultShouldWait(waitMs uint64) *TokenResult {
	result := resultPool.Get().(*TokenResult)
	result.status = ResultStatusShouldWait
	result.waitMs = waitMs
	return result
}

func RefurbishTokenResult(result *TokenResult) {
	if result != nil {
		result.status = ResultStatusPass
		result.blockErr = nil
		result.waitMs = 0

		resultPool.Put(result)
	}
}
