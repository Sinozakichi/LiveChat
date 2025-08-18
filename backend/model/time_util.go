package model

import "time"

var timeNow = time.Now

// getCurrentTimestamp 獲取當前時間戳（秒）
func getCurrentTimestamp() int64 {
	return timeNow().Unix()
}

// SetTimeNow 設置時間函數，用於測試
func SetTimeNow(f func() time.Time) {
	timeNow = f
}

// ResetTimeNow 重置時間函數為默認值
func ResetTimeNow() {
	timeNow = time.Now
}
