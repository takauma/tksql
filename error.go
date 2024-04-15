package tksql

// NoResultError 結果なしエラー構造体.
type NoResultError struct {
	message string
}

// Error エラー内容を返します.
func (e *NoResultError) Error() string {
	return e.message
}

// NewNoResultError 結果なしエラー構造体インスタンスを生成します.
func NewNoResultError(message string) *NoResultError {
	return &NoResultError{message}
}
