package resume

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// ========== Benchmark Helper Functions ==========

// createBenchmarkTask creates a mock task record for benchmarking.
func createBenchmarkTask(title, priority, description string) *mockRecord {
	return newMockRecord("abc123def456", map[string]any{
		"title":       title,
		"priority":    priority,
		"description": description,
		"seq":         42,
	})
}

// createBenchmarkComments creates a slice of comments for benchmarking.
func createBenchmarkComments(count int) []Comment {
	comments := make([]Comment, count)
	baseTime := time.Now()

	for i := 0; i < count; i++ {
		authorType := "human"
		authorId := "user"
		if i%2 == 0 {
			authorType = "agent"
			authorId = "opencode"
		}

		comments[i] = Comment{
			Content:    fmt.Sprintf("This is benchmark comment %d with some realistic content to simulate actual usage patterns", i),
			AuthorType: authorType,
			AuthorId:   authorId,
			Created:    baseTime.Add(time.Duration(i) * time.Minute),
		}
	}

	return comments
}

// createLargeComment creates a comment with specified content size.
func createLargeComment(size int) Comment {
	content := make([]byte, size)
	for i := range content {
		content[i] = byte('a' + (i % 26))
	}
	return Comment{
		Content:    string(content),
		AuthorType: "human",
		AuthorId:   "user",
		Created:    time.Now(),
	}
}

// ========== Core Operation Benchmarks ==========

// BenchmarkBuildContextPrompt measures the performance of building a full context prompt.
// Target: < 50ms
// Tests: full prompt generation with typical comment count
func BenchmarkBuildContextPrompt(b *testing.B) {
	task := createBenchmarkTask(
		"Implement user authentication",
		"high",
		"We need to add user authentication to the application using JWT tokens.",
	)
	comments := createBenchmarkComments(10)
	displayId := "WRK-42"

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buildContextPromptFromMock(task, comments)
		_ = displayId // displayId would be used in actual call
	}
}

// BenchmarkBuildMinimalPrompt measures the performance of building a minimal prompt.
// Target: < 50ms
// Tests: minimal prompt generation (last 3 comments only)
func BenchmarkBuildMinimalPrompt(b *testing.B) {
	task := createBenchmarkTask(
		"Implement user authentication",
		"high",
		"",
	)
	comments := createBenchmarkComments(10)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buildMinimalPromptFromMock(task, comments)
	}
}

// BenchmarkFormatAuthorLabel measures author label formatting performance.
func BenchmarkFormatAuthorLabel(b *testing.B) {
	testCases := []struct {
		name       string
		authorType string
		authorId   string
	}{
		{"with_id", "agent", "opencode"},
		{"without_id", "human", ""},
		{"long_id", "agent", "very-long-agent-name-for-testing"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				formatAuthorLabel(tc.authorType, tc.authorId)
			}
		})
	}
}

// ========== Scaling Benchmarks ==========

// BenchmarkContextPromptScaling measures how prompt building scales with comment count.
// Target: Handle 100+ comments efficiently
func BenchmarkContextPromptScaling(b *testing.B) {
	commentCounts := []int{1, 10, 50, 100, 200}

	for _, count := range commentCounts {
		b.Run(fmt.Sprintf("comments_%d", count), func(b *testing.B) {
			task := createBenchmarkTask(
				"Implement feature",
				"medium",
				"Feature description here.",
			)
			comments := createBenchmarkComments(count)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buildContextPromptFromMock(task, comments)
			}
		})
	}
}

// BenchmarkMinimalPromptScaling measures minimal prompt with varying comment counts.
// The minimal prompt only uses last 3 comments, so this should be constant time.
func BenchmarkMinimalPromptScaling(b *testing.B) {
	commentCounts := []int{1, 10, 50, 100, 200}

	for _, count := range commentCounts {
		b.Run(fmt.Sprintf("comments_%d", count), func(b *testing.B) {
			task := createBenchmarkTask(
				"Implement feature",
				"medium",
				"",
			)
			comments := createBenchmarkComments(count)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buildMinimalPromptFromMock(task, comments)
			}
		})
	}
}

// BenchmarkLargeCommentContent measures performance with large comment content.
func BenchmarkLargeCommentContent(b *testing.B) {
	contentSizes := []int{100, 500, 1000, 2000}

	for _, size := range contentSizes {
		b.Run(fmt.Sprintf("content_%d_chars", size), func(b *testing.B) {
			task := createBenchmarkTask(
				"Task with large comments",
				"high",
				"",
			)
			comments := []Comment{createLargeComment(size)}

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buildContextPromptFromMock(task, comments)
			}
		})
	}
}

// BenchmarkDescriptionTruncation measures performance with long descriptions.
func BenchmarkDescriptionTruncation(b *testing.B) {
	descriptionSizes := []int{100, 500, 1000}

	for _, size := range descriptionSizes {
		b.Run(fmt.Sprintf("description_%d_chars", size), func(b *testing.B) {
			description := strings.Repeat("A", size)
			task := createBenchmarkTask(
				"Task with long description",
				"high",
				description,
			)
			comments := createBenchmarkComments(5)

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				buildContextPromptFromMock(task, comments)
			}
		})
	}
}

// ========== Memory Allocation Benchmarks ==========

// BenchmarkContextPromptAllocs measures memory allocations during prompt building.
// Run with: go test -bench=BenchmarkContextPromptAllocs -benchmem
func BenchmarkContextPromptAllocs(b *testing.B) {
	task := createBenchmarkTask(
		"Implement user authentication",
		"high",
		"We need to add user authentication to the application.",
	)
	comments := createBenchmarkComments(50)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buildContextPromptFromMock(task, comments)
	}
}

// BenchmarkMinimalPromptAllocs measures memory allocations for minimal prompts.
// Run with: go test -bench=BenchmarkMinimalPromptAllocs -benchmem
func BenchmarkMinimalPromptAllocs(b *testing.B) {
	task := createBenchmarkTask(
		"Implement user authentication",
		"high",
		"",
	)
	comments := createBenchmarkComments(50)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buildMinimalPromptFromMock(task, comments)
	}
}

// ========== String Builder Efficiency Benchmarks ==========

// BenchmarkStringBuilderVsConcat compares string builder with naive concatenation.
// This validates that our current implementation using strings.Builder is efficient.
func BenchmarkStringBuilderVsConcat(b *testing.B) {
	comments := createBenchmarkComments(20)

	b.Run("string_builder", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var sb strings.Builder
			for _, c := range comments {
				sb.WriteString("[")
				sb.WriteString(c.AuthorId)
				sb.WriteString("]: ")
				sb.WriteString(c.Content)
				sb.WriteString("\n")
			}
			_ = sb.String()
		}
	})

	b.Run("naive_concat", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			result := ""
			for _, c := range comments {
				result += "[" + c.AuthorId + "]: " + c.Content + "\n"
			}
			_ = result
		}
	})
}

// BenchmarkGetDisplayId measures display ID generation performance.
func BenchmarkGetDisplayId(b *testing.B) {
	testCases := []struct {
		name string
		task *mockRecord
	}{
		{
			"with_display_id",
			newMockRecord("abc123def456", map[string]any{"display_id": "PROJ-99", "seq": 42}),
		},
		{
			"with_seq",
			newMockRecord("abc123def456", map[string]any{"seq": 42}),
		},
		{
			"fallback_to_id",
			newMockRecord("abc123def456ghij", map[string]any{}),
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				getDisplayIdFromMock(tc.task)
			}
		})
	}
}

// BenchmarkCommentTimeFormatting measures time formatting overhead.
func BenchmarkCommentTimeFormatting(b *testing.B) {
	now := time.Now()

	b.Run("format_15_04", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = now.Format("15:04")
		}
	})

	b.Run("format_RFC3339", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = now.Format(time.RFC3339)
		}
	})
}

// ========== Realistic Scenario Benchmarks ==========

// BenchmarkRealisticResumeScenario simulates a realistic resume scenario.
// This represents the full context building that happens during task resume.
func BenchmarkRealisticResumeScenario(b *testing.B) {
	// Realistic scenario: task with detailed description and conversation
	task := createBenchmarkTask(
		"Implement user authentication with OAuth2",
		"high",
		`We need to implement OAuth2 authentication for our application.
Requirements:
- Support Google and GitHub as OAuth providers
- Implement refresh token rotation
- Store tokens securely in the database
- Add logout functionality that revokes tokens`,
	)

	// Realistic conversation thread
	comments := []Comment{
		{
			Content:    "I've started working on the OAuth2 implementation. Should I use the official Go OAuth2 library or implement it from scratch?",
			AuthorType: "agent",
			AuthorId:   "opencode",
			Created:    time.Now().Add(-4 * time.Hour),
		},
		{
			Content:    "@agent Use the official golang.org/x/oauth2 package. It's well-maintained and handles the OAuth2 flow correctly.",
			AuthorType: "human",
			AuthorId:   "senior-dev",
			Created:    time.Now().Add(-3 * time.Hour),
		},
		{
			Content:    "Got it. I've implemented the basic OAuth2 flow with Google. For token storage, should I use encrypted cookies or store them in the database?",
			AuthorType: "agent",
			AuthorId:   "opencode",
			Created:    time.Now().Add(-2 * time.Hour),
		},
		{
			Content:    "@agent Store tokens in the database. Use the existing sessions table and encrypt the tokens at rest. Access tokens in memory only.",
			AuthorType: "human",
			AuthorId:   "senior-dev",
			Created:    time.Now().Add(-1 * time.Hour),
		},
		{
			Content:    "I need clarification on the refresh token rotation. Should we issue new refresh tokens on every access token refresh, or only when the refresh token is close to expiry?",
			AuthorType: "agent",
			AuthorId:   "opencode",
			Created:    time.Now(),
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Build full context (used for full resume)
		buildContextPromptFromMock(task, comments)

		// Build minimal context (used for quick resume)
		buildMinimalPromptFromMock(task, comments)
	}
}

// BenchmarkEmptyConversation measures performance with no comments.
func BenchmarkEmptyConversation(b *testing.B) {
	task := createBenchmarkTask(
		"New task with no comments",
		"medium",
		"This is a new task that hasn't had any conversation yet.",
	)
	comments := []Comment{}

	b.Run("context_prompt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buildContextPromptFromMock(task, comments)
		}
	})

	b.Run("minimal_prompt", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			buildMinimalPromptFromMock(task, comments)
		}
	})
}
