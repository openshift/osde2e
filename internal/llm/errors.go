package llm

import "errors"

var (
	ErrNoResponseCandidates = errors.New("no response candidates from gemini")
	ErrNoContentInResponse  = errors.New("no content in gemini response")
	ErrToolCallFailed       = errors.New("failed to handle tool call")
	ErrMaxIterations        = errors.New("max iterations reached without final response")
)
