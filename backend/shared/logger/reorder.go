package logger

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/rs/zerolog"
)

// orderedWriter wraps an io.Writer and rewrites every JSON log line so that
// fields appear in a predictable order:
//
//  1. Priority fields in the declared sequence (missing ones are skipped).
//  2. Any remaining fields preserving their original insertion order.
//  3. The "message" field always last.
//
// This is necessary because zerolog appends Timestamp() and Caller() via hooks
// that run after all explicit event fields, making their position in the JSON
// output non-deterministic relative to user-set fields like "id" or "method".
type orderedWriter struct {
	w        io.Writer
	priority []string
}

// Write implements io.Writer.
func (ow *orderedWriter) Write(p []byte) (int, error) {
	reordered := reorderJSONLine(p, ow.priority)
	if _, err := ow.w.Write(reordered); err != nil {
		return 0, err
	}
	// Return the original length so zerolog never sees a "short write" error.
	return len(p), nil
}

// WriteLevel implements zerolog.LevelWriter so the wrapper integrates
// seamlessly with zerolog.New().
func (ow *orderedWriter) WriteLevel(_ zerolog.Level, p []byte) (int, error) {
	return ow.Write(p)
}

// reorderJSONLine parses a single JSON log line, reorders its fields, and
// re-serialises it. If the line is not valid JSON it is returned unchanged.
func reorderJSONLine(line []byte, priority []string) []byte {
	trimmed := bytes.TrimRight(line, "\n")

	dec := json.NewDecoder(bytes.NewReader(trimmed))
	dec.UseNumber()

	// Read opening '{'.
	if tok, err := dec.Token(); err != nil || tok != json.Delim('{') {
		return line
	}

	type kv struct {
		key string
		val json.RawMessage
	}

	// Collect all key-value pairs in original order.
	var pairs []kv
	index := map[string]json.RawMessage{}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return line
		}
		key, ok := keyTok.(string)
		if !ok {
			return line
		}
		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			return line
		}
		pairs = append(pairs, kv{key, val})
		index[key] = val
	}

	msgKey := zerolog.MessageFieldName // "message"

	var buf bytes.Buffer
	buf.WriteByte('{')
	written := 0

	writeKV := func(key string, val json.RawMessage) {
		if written > 0 {
			buf.WriteByte(',')
		}
		written++
		keyBytes, _ := json.Marshal(key)
		buf.Write(keyBytes)
		buf.WriteByte(':')
		buf.Write(val)
	}

	// 1. Priority fields in declared order.
	for _, key := range priority {
		val, ok := index[key]
		if !ok {
			continue
		}
		writeKV(key, val)
		delete(index, key)
	}

	// 2. Remaining fields in original insertion order, except "message".
	for _, p := range pairs {
		if p.key == msgKey {
			continue
		}
		if _, stillPending := index[p.key]; !stillPending {
			continue // already written as a priority field
		}
		writeKV(p.key, p.val)
		delete(index, p.key)
	}

	// 3. "message" always last.
	for _, p := range pairs {
		if p.key == msgKey {
			writeKV(p.key, p.val)
			break
		}
	}

	buf.WriteByte('}')
	buf.WriteByte('\n')
	return buf.Bytes()
}
