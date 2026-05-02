package main

import (
	"strings"
	"testing"
)

func BenchmarkOld(b *testing.B) {
	streamItems := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		streamItems[i] = `{"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-3.5-turbo-0125","system_fingerprint":"fp_44709d6fcb","choices":[{"index":0,"delta":{"content":"Hello"},"logprobs":null,"finish_reason":null}]}`
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = "[" + strings.Join(streamItems, ",") + "]"
	}
}

func BenchmarkNew(b *testing.B) {
	streamItems := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		streamItems[i] = `{"id":"chatcmpl-123","object":"chat.completion.chunk","created":1694268190,"model":"gpt-3.5-turbo-0125","system_fingerprint":"fp_44709d6fcb","choices":[{"index":0,"delta":{"content":"Hello"},"logprobs":null,"finish_reason":null}]}`
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		capacity := 2
		for _, item := range streamItems {
			capacity += len(item) + 1
		}

		streamResp := make([]byte, 0, capacity)
		streamResp = append(streamResp, '[')
		for i, item := range streamItems {
			if i > 0 {
				streamResp = append(streamResp, ',')
			}
			streamResp = append(streamResp, item...)
		}
		streamResp = append(streamResp, ']')
		_ = string(streamResp) // Since we need to pass a string eventually
	}
}
