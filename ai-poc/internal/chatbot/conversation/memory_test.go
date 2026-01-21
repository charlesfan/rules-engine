package conversation

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/charlesfan/rules-engine/ai-poc/pkg/llm"
)

// mockCompressor is a mock compressor for testing
type mockCompressor struct {
	compressFunc func(ctx context.Context, existingSummary string, messages []llm.Message) (string, error)
	callCount    int
}

func (m *mockCompressor) Compress(ctx context.Context, existingSummary string, messages []llm.Message) (string, error) {
	m.callCount++
	if m.compressFunc != nil {
		return m.compressFunc(ctx, existingSummary, messages)
	}
	return "mock summary", nil
}

func TestMemory_AddMessages(t *testing.T) {
	mem := NewMemory()

	mem.AddUserMessage("Hello")
	mem.AddAssistantMessage("Hi there")

	messages := mem.GetMessages()
	assert.Len(t, messages, 2)
	assert.Equal(t, llm.RoleUser, messages[0].Role)
	assert.Equal(t, "Hello", messages[0].Content)
	assert.Equal(t, llm.RoleAssistant, messages[1].Role)
	assert.Equal(t, "Hi there", messages[1].Content)
}

func TestMemory_Clear(t *testing.T) {
	mem := NewMemory()

	mem.AddUserMessage("Hello")
	mem.AddAssistantMessage("Hi")
	assert.Equal(t, 2, mem.Len())

	mem.Clear()
	assert.Equal(t, 0, mem.Len())
	assert.Empty(t, mem.GetSummary())
}

func TestMemory_NeedsCompression(t *testing.T) {
	config := MemoryConfig{
		CompressThreshold: 4,
		KeepAfterCompress: 2,
	}
	mem := NewMemoryWithConfig(config)

	// Add 3 messages - should not need compression
	mem.AddUserMessage("1")
	mem.AddAssistantMessage("2")
	mem.AddUserMessage("3")
	assert.False(t, mem.NeedsCompression())

	// Add 1 more - now at threshold
	mem.AddAssistantMessage("4")
	assert.True(t, mem.NeedsCompression())
}

func TestMemory_Compress(t *testing.T) {
	config := MemoryConfig{
		CompressThreshold: 4,
		KeepAfterCompress: 2,
	}
	mem := NewMemoryWithConfig(config)

	compressor := &mockCompressor{
		compressFunc: func(ctx context.Context, existingSummary string, messages []llm.Message) (string, error) {
			// Should receive 2 messages (4 - 2 = 2 to compress)
			assert.Len(t, messages, 2)
			return "compressed: " + messages[0].Content + ", " + messages[1].Content, nil
		},
	}
	mem.SetCompressor(compressor)

	// Add 4 messages
	mem.AddUserMessage("msg1")
	mem.AddAssistantMessage("msg2")
	mem.AddUserMessage("msg3")
	mem.AddAssistantMessage("msg4")

	// Compress
	err := mem.Compress(context.Background())
	require.NoError(t, err)

	// Check results
	assert.Equal(t, "compressed: msg1, msg2", mem.GetSummary())
	assert.Equal(t, 2, mem.Len()) // Only 2 messages kept

	messages := mem.GetMessages()
	assert.Equal(t, "msg3", messages[0].Content)
	assert.Equal(t, "msg4", messages[1].Content)

	// Compressor should be called once
	assert.Equal(t, 1, compressor.callCount)
}

func TestMemory_Compress_NoCompressor(t *testing.T) {
	mem := NewMemory()
	mem.AddUserMessage("test")

	err := mem.Compress(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no compressor set")
}

func TestMemory_TryCompress(t *testing.T) {
	config := MemoryConfig{
		CompressThreshold: 2,
		KeepAfterCompress: 1,
	}
	mem := NewMemoryWithConfig(config)

	compressor := &mockCompressor{}
	mem.SetCompressor(compressor)

	// Add 1 message - should not trigger compression
	mem.AddUserMessage("msg1")
	err := mem.TryCompress(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 0, compressor.callCount)

	// Add 1 more - should trigger compression
	mem.AddAssistantMessage("msg2")
	err = mem.TryCompress(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 1, compressor.callCount)
}

func TestMemory_BuildRequest(t *testing.T) {
	mem := NewMemory()

	// Set a summary manually for testing
	mem.mu.Lock()
	mem.summary = "之前討論了早鳥優惠"
	mem.mu.Unlock()

	mem.AddUserMessage("團報優惠呢？")
	mem.AddAssistantMessage("團報優惠是...")

	req := mem.BuildRequest("你是助手")

	// Should have: system prompt + summary + 2 messages = 4 messages
	assert.Len(t, req.Messages, 4)
	assert.Equal(t, llm.RoleSystem, req.Messages[0].Role)
	assert.Equal(t, "你是助手", req.Messages[0].Content)
	assert.Equal(t, llm.RoleSystem, req.Messages[1].Role)
	assert.Contains(t, req.Messages[1].Content, "對話摘要")
	assert.Contains(t, req.Messages[1].Content, "早鳥優惠")
}

func TestMemory_BuildRequest_NoSummary(t *testing.T) {
	mem := NewMemory()

	mem.AddUserMessage("Hello")

	req := mem.BuildRequest("System prompt")

	// Should have: system prompt + 1 message = 2 messages
	assert.Len(t, req.Messages, 2)
}

func TestMemory_Metadata(t *testing.T) {
	mem := NewMemory()

	mem.SetMetadata("event_type", "marathon")
	mem.SetMetadata("count", 42)

	val, ok := mem.GetMetadata("event_type")
	assert.True(t, ok)
	assert.Equal(t, "marathon", val)

	val, ok = mem.GetMetadata("count")
	assert.True(t, ok)
	assert.Equal(t, 42, val)

	_, ok = mem.GetMetadata("nonexistent")
	assert.False(t, ok)
}

func TestMemory_Stats(t *testing.T) {
	config := MemoryConfig{
		CompressThreshold: 2,
		KeepAfterCompress: 1,
	}
	mem := NewMemoryWithConfig(config)
	mem.SetCompressor(&mockCompressor{})

	mem.AddUserMessage("1")
	mem.AddAssistantMessage("2")
	_ = mem.TryCompress(context.Background())

	stats := mem.GetStats()
	assert.Equal(t, 1, stats.RecentMessageCount)
	assert.True(t, stats.HasSummary)
	assert.Equal(t, 1, stats.CompressCount)
}

func TestMemory_TotalLen(t *testing.T) {
	config := MemoryConfig{
		CompressThreshold: 4,
		KeepAfterCompress: 2,
	}
	mem := NewMemoryWithConfig(config)
	mem.SetCompressor(&mockCompressor{})

	// Add 4 messages and compress
	for i := 0; i < 4; i++ {
		mem.AddUserMessage("msg")
	}
	_ = mem.Compress(context.Background())

	// Add 2 more
	mem.AddUserMessage("new1")
	mem.AddUserMessage("new2")

	// Total should be: 2 (compressed) + 2 (kept) + 2 (new) = 6
	// But our estimate is: 1 * (4-2) + 4 = 6
	assert.Equal(t, 6, mem.TotalLen())
}

func TestSimpleCompressor(t *testing.T) {
	comp := NewSimpleCompressor(200)

	messages := []llm.Message{
		llm.NewUserMessage("設定早鳥優惠"),
		llm.NewAssistantMessage("好的"),
		llm.NewUserMessage("8折"),
	}

	result, err := comp.Compress(context.Background(), "", messages)
	require.NoError(t, err)
	assert.Contains(t, result, "早鳥優惠")
	assert.Contains(t, result, "8折")
}

func TestSimpleCompressor_WithExistingSummary(t *testing.T) {
	comp := NewSimpleCompressor(500)

	messages := []llm.Message{
		llm.NewUserMessage("團報優惠"),
	}

	result, err := comp.Compress(context.Background(), "已設定早鳥優惠8折", messages)
	require.NoError(t, err)
	assert.Contains(t, result, "早鳥優惠8折")
	assert.Contains(t, result, "團報優惠")
}

func TestNoOpCompressor(t *testing.T) {
	comp := NewNoOpCompressor()

	messages := []llm.Message{
		llm.NewUserMessage("test"),
	}

	result, err := comp.Compress(context.Background(), "existing", messages)
	require.NoError(t, err)
	assert.Equal(t, "existing", result)
}
