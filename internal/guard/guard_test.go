package guard

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jjmrocha/ai-toolkit/classify"
	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRepoPath = "/src/joe"
	testKBPath   = "/srv/wiki"
)

var errUnavailable = errors.New("classifier unavailable")

type answerFunc func(ctx context.Context, req classify.Request) (*classify.Response, error)

type fakeClassifier struct {
	answer   answerFunc
	requests []classify.Request
}

func (f *fakeClassifier) Classify(ctx context.Context, req classify.Request) (*classify.Response, error) {
	f.requests = append(f.requests, req)

	return f.answer(ctx, req)
}

func answeringWith(answer classify.Answer) *fakeClassifier {
	return &fakeClassifier{answer: func(_ context.Context, req classify.Request) (*classify.Response, error) {
		answers := map[string]classify.Answer{}
		for id := range req.Questions {
			answers[id] = answer
		}

		return &classify.Response{Answers: answers}, nil
	}}
}

func answering(value float64) *fakeClassifier {
	return answeringWith(classify.YesNoAnswer{Value: value})
}

func scripted(values map[string]float64) *fakeClassifier {
	return &fakeClassifier{answer: func(_ context.Context, req classify.Request) (*classify.Response, error) {
		answers := map[string]classify.Answer{}

		for id := range req.Questions {
			if value, ok := values[id]; ok {
				answers[id] = classify.YesNoAnswer{Value: value}
			}
		}

		return &classify.Response{Answers: answers}, nil
	}}
}

func failingFirst(fake *fakeClassifier) *fakeClassifier {
	answer := fake.answer
	failed := false

	fake.answer = func(ctx context.Context, req classify.Request) (*classify.Response, error) {
		if !failed {
			failed = true

			return nil, errUnavailable
		}

		return answer(ctx, req)
	}

	return fake
}

func questionIDs(req classify.Request) []string {
	return slices.Sorted(maps.Keys(req.Questions))
}

func newInterceptor(t *testing.T, cfg Config) tools.Interceptor {
	t.Helper()

	interceptor, err := NewInterceptor(cfg)
	require.NoError(t, err)

	return interceptor
}

func shellBox(t *testing.T) *tools.ToolBox {
	t.Helper()

	toolBox := tools.NewToolBox()
	tool := llm.Tool{
		Name:        "shell_run",
		Description: "Run a command line",
		Schema:      map[string]any{},
	}
	require.NoError(t, toolBox.Add(tool, func(context.Context, map[string]any) (string, error) { return "", nil }))

	return toolBox
}

func shellCall(command string) llm.ToolCall {
	return llm.ToolCall{ID: "call-1", Name: "shell_run", Arguments: map[string]any{"command": command}}
}

func shellCallOfInputSize(t *testing.T, size int) llm.ToolCall {
	t.Helper()

	tool, found := shellBox(t).Tool("shell_run")
	require.True(t, found)

	input, err := classifierInput(tool, shellCall(""), Constraints(testRepoPath, testKBPath))
	require.NoError(t, err)

	return shellCall(strings.Repeat("a", size-len(input)))
}

func TestNewInterceptor(t *testing.T) {
	t.Run("blocks a call judged to violate the constraints", func(t *testing.T) {
		// given
		interceptor := newInterceptor(t, Config{Classifier: answering(0.93), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		// when
		result := interceptor(t.Context(), shellCall("git push --force origin main"))
		// then
		require.ErrorIs(t, result, ErrToolCallRejected)
		assert.Contains(t, result.Error(), "p=0.93")
	})

	t.Run("blocks a call judged exactly at the threshold", func(t *testing.T) {
		// given
		interceptor := newInterceptor(t, Config{Classifier: answering(0.8), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		// when
		result := interceptor(t.Context(), shellCall("rm -rf ~/Documents"))
		// then
		assert.ErrorIs(t, result, ErrToolCallRejected)
	})

	t.Run("allows a call judged below the threshold", func(t *testing.T) {
		// given
		interceptor := newInterceptor(t, Config{Classifier: answering(0.79), ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		// when
		result := interceptor(t.Context(), shellCall("go test ./..."))
		// then
		assert.NoError(t, result)
	})

	t.Run("allows the call when the classifier cannot answer", func(t *testing.T) {
		testCases := []struct {
			name       string
			classifier *fakeClassifier
		}{
			{
				name:       "classifier error",
				classifier: failingFirst(answering(0.93)),
			},
			{
				name:       "missing answer",
				classifier: scripted(map[string]float64{}),
			},
			{
				name:       "answer of another type",
				classifier: answeringWith(classify.ChoiceAnswer{Selected: "yes", Probabilities: map[string]float64{"yes": 1}}),
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				interceptor := newInterceptor(t, Config{Classifier: testCase.classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
				// when
				result := interceptor(t.Context(), shellCall("git push --force origin main"))
				// then
				assert.NoError(t, result)
			})
		}
	})

	t.Run("allows the call when the caller gives up before the classifier answers", func(t *testing.T) {
		// given
		hanging := &fakeClassifier{answer: func(ctx context.Context, _ classify.Request) (*classify.Response, error) {
			<-ctx.Done()

			return nil, ctx.Err()
		}}
		interceptor := newInterceptor(t, Config{Classifier: hanging, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})

		ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
		defer cancel()
		// when
		result := interceptor(ctx, shellCall("git push --force origin main"))
		// then
		assert.NoError(t, result)
	})

	t.Run("allows a call to a tool the box does not hold without asking", func(t *testing.T) {
		// given
		classifier := answering(0.1)
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		call := llm.ToolCall{ID: "call-1", Name: "file_delete", Arguments: map[string]any{"path": "/etc/hosts"}}
		// when
		result := interceptor(t.Context(), call)
		// then
		require.NoError(t, result)
		assert.Empty(t, classifier.requests)
	})

	t.Run("asks whether the call and the tool violate the constraints on the tool's first call", func(t *testing.T) {
		// given
		classifier := answering(0.1)
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		expected := map[string]classify.Question{
			callQuestionID: classify.YesNo{Instructions: "Does this tool call violate any of the constraints?"},
			toolQuestionID: classify.YesNo{Instructions: "Could any call to this tool, with any arguments, violate any of the constraints?"},
		}
		// when
		err := interceptor(t.Context(), shellCall("go test ./..."))
		// then
		require.NoError(t, err)
		require.Len(t, classifier.requests, 1)
		assert.Equal(t, expected, classifier.requests[0].Questions)
	})

	t.Run("skips the classifier for a tool judged safe", func(t *testing.T) {
		// given
		classifier := scripted(map[string]float64{callQuestionID: 0.1, toolQuestionID: 0.19})
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		require.NoError(t, interceptor(t.Context(), shellCall("date")))
		// when
		result := interceptor(t.Context(), shellCall("rm -rf ~/Documents"))
		// then
		require.NoError(t, result)
		assert.Len(t, classifier.requests, 1)
	})

	t.Run("asks only about the call once the tool is judged unsafe", func(t *testing.T) {
		testCases := []struct {
			name     string
			toolSafe float64
		}{
			{name: "at the safe threshold", toolSafe: 0.2},
			{name: "well above the safe threshold", toolSafe: 0.83},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				classifier := scripted(map[string]float64{callQuestionID: 0.1, toolQuestionID: testCase.toolSafe})
				interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
				require.NoError(t, interceptor(t.Context(), shellCall("go build ./...")))

				expected := []string{callQuestionID}
				// when
				err := interceptor(t.Context(), shellCall("go test ./..."))
				// then
				require.NoError(t, err)
				require.Len(t, classifier.requests, 2)
				assert.Equal(t, expected, questionIDs(classifier.requests[1]))
			})
		}
	})

	t.Run("asks only about the call once a call to a tool judged safe was blocked", func(t *testing.T) {
		// given
		classifier := scripted(map[string]float64{callQuestionID: 0.93, toolQuestionID: 0.05})
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		require.ErrorIs(t, interceptor(t.Context(), shellCall("git push --force origin main")), ErrToolCallRejected)

		expected := []string{callQuestionID}
		// when
		err := interceptor(t.Context(), shellCall("git push --force origin main"))
		// then
		require.ErrorIs(t, err, ErrToolCallRejected)
		require.Len(t, classifier.requests, 2)
		assert.Equal(t, expected, questionIDs(classifier.requests[1]))
	})

	t.Run("asks about the tool again when it could not be classified", func(t *testing.T) {
		testCases := []struct {
			name       string
			classifier *fakeClassifier
		}{
			{
				name:       "classifier error",
				classifier: failingFirst(scripted(map[string]float64{callQuestionID: 0.1, toolQuestionID: 0.05})),
			},
			{
				name:       "missing tool answer",
				classifier: scripted(map[string]float64{callQuestionID: 0.1}),
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				classifier := testCase.classifier
				interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
				require.NoError(t, interceptor(t.Context(), shellCall("go build ./...")))

				expected := []string{toolQuestionID, callQuestionID}
				// when
				err := interceptor(t.Context(), shellCall("go test ./..."))
				// then
				require.NoError(t, err)
				require.NotEmpty(t, classifier.requests)
				assert.Equal(t, expected, questionIDs(classifier.requests[len(classifier.requests)-1]))
			})
		}
	})

	t.Run("describes the tool, the call and the constraints", func(t *testing.T) {
		testCases := []struct {
			name     string
			call     llm.ToolCall
			kbPath   string
			expected string
		}{
			{
				name:   "with a knowledge base",
				call:   shellCall("go test ./..."),
				kbPath: testKBPath,
				expected: `Tool: {"name":"shell_run","description":"Run a command line","schema":{}}
Call: {"name":"shell_run","arguments":{"command":"go test ./..."}}
Constraints:
- The agent may only create, modify or delete files inside /src/joe.
- The agent may also create, modify or delete files inside /srv/wiki.
- The agent must not change remote or shared state: push, deploy, publish, merge, or send messages.
- The agent must not read or transmit secrets or credentials outside the machine.`,
			},
			{
				name: "without a knowledge base",
				call: shellCall("go test ./..."),
				expected: `Tool: {"name":"shell_run","description":"Run a command line","schema":{}}
Call: {"name":"shell_run","arguments":{"command":"go test ./..."}}
Constraints:
- The agent may only create, modify or delete files inside /src/joe.
- The agent must not change remote or shared state: push, deploy, publish, merge, or send messages.
- The agent must not read or transmit secrets or credentials outside the machine.`,
			},
			{
				name: "keeps shell operators readable",
				call: shellCall("go build ./... && go test ./... > out.txt"),
				expected: `Tool: {"name":"shell_run","description":"Run a command line","schema":{}}
Call: {"name":"shell_run","arguments":{"command":"go build ./... && go test ./... > out.txt"}}
Constraints:
- The agent may only create, modify or delete files inside /src/joe.
- The agent must not change remote or shared state: push, deploy, publish, merge, or send messages.
- The agent must not read or transmit secrets or credentials outside the machine.`,
			},
		}

		for _, testCase := range testCases {
			t.Run(testCase.name, func(t *testing.T) {
				// given
				classifier := answering(0.1)
				interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testCase.kbPath})
				// when
				err := interceptor(t.Context(), testCase.call)
				// then
				require.NoError(t, err)
				require.Len(t, classifier.requests, 1)
				assert.Equal(t, testCase.expected, classifier.requests[0].Input)
			})
		}
	})

	t.Run("allows a call whose arguments cannot be described without asking", func(t *testing.T) {
		// given
		classifier := answering(0.93)
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		call := shellCall("")
		call.Arguments["command"] = make(chan int)
		// when
		result := interceptor(t.Context(), call)
		// then
		require.NoError(t, result)
		assert.Empty(t, classifier.requests)
	})

	t.Run("rejects a call too big to classify without asking", func(t *testing.T) {
		// given
		classifier := answering(0.1)
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		// when
		result := interceptor(t.Context(), shellCallOfInputSize(t, maxInputBytes+1))
		// then
		require.ErrorIs(t, result, ErrToolCallTooLarge)
		assert.Empty(t, classifier.requests)
	})

	t.Run("classifies a call exactly at the size limit", func(t *testing.T) {
		// given
		classifier := answering(0.1)
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		// when
		result := interceptor(t.Context(), shellCallOfInputSize(t, maxInputBytes))
		// then
		require.NoError(t, result)
		require.Len(t, classifier.requests, 1)
		assert.Len(t, classifier.requests[0].Input, maxInputBytes)
	})

	t.Run("allows a call too big to classify to a tool judged safe", func(t *testing.T) {
		// given
		classifier := scripted(map[string]float64{callQuestionID: 0.1, toolQuestionID: 0.1})
		interceptor := newInterceptor(t, Config{Classifier: classifier, ToolBox: shellBox(t), RepoPath: testRepoPath, KBPath: testKBPath})
		require.NoError(t, interceptor(t.Context(), shellCall("date")))
		// when
		result := interceptor(t.Context(), shellCallOfInputSize(t, maxInputBytes+1))
		// then
		require.NoError(t, result)
		assert.Len(t, classifier.requests, 1)
	})
}
