package session

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/jjmrocha/ai-toolkit/llm"
	"github.com/jjmrocha/ai-toolkit/subagent"
	"github.com/jjmrocha/joe/internal/config"
)

type SessionIDSource interface {
	SessionID() string
}

type debugWriter struct {
	dir      string
	repoPath string
	main     SessionIDSource
	now      func() time.Time
	mu       sync.Mutex
	last     map[string]int
}

func Debugger(repoPath string, main SessionIDSource) (subagent.Debugger, error) {
	return newDebugger(repoPath, main, time.Now)
}

func newDebugger(repoPath string, main SessionIDSource, now func() time.Time) (subagent.Debugger, error) {
	dir, err := config.DebugDir()
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}

	w := &debugWriter{
		dir:      dir,
		repoPath: repoPath,
		main:     main,
		now:      now,
		last:     make(map[string]int),
	}

	return w.save, nil
}

func (w *debugWriter) save(agentName string, messages []llm.Message, runErr error) {
	id := w.main.SessionID()

	run := debugRun{
		Session:  id,
		Sequence: w.next(id),
		Agent:    agentName,
		Finished: w.now().UTC().Truncate(time.Second),
		Repo:     w.repoPath,
		Messages: toMessages(messages),
	}

	if runErr != nil {
		run.Error = runErr.Error()
	}

	_ = writeJSON(w.dir, fmt.Sprintf("%s-%03d-%s", id, run.Sequence, agentName), &run)
}

func (w *debugWriter) next(id string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	last, ok := w.last[id]
	if !ok {
		last = w.highestSaved(id)
	}

	w.last[id] = last + 1

	return last + 1
}

func (w *debugWriter) highestSaved(id string) int {
	entries, err := os.ReadDir(w.dir)
	if err != nil {
		return 0
	}

	highest := 0

	for _, entry := range entries {
		if n := sequenceOf(entry.Name(), id); n > highest {
			highest = n
		}
	}

	return highest
}

func sequenceOf(name, id string) int {
	rest, ok := strings.CutPrefix(name, id+"-")
	if !ok {
		return 0
	}

	rest, ok = strings.CutSuffix(rest, ".json")
	if !ok {
		return 0
	}

	digits, agentName, ok := strings.Cut(rest, "-")
	if !ok || agentName == "" {
		return 0
	}

	n, err := strconv.Atoi(digits)
	if err != nil {
		return 0
	}

	return n
}
