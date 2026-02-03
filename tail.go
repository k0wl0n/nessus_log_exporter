package main

import (
	"bufio"
	"fmt"
	"os"
)

// tailLines opens path, seeks to the saved offset in state, reads every new
// line, calls fn(line), then persists the new offset.
// If the file has been rotated (truncated / smaller than cursor) the cursor
// resets to 0 automatically.
func tailLines(path string, state *exporterState, fn func(string)) error {
	f, err := os.Open(path)
	if err != nil {
		// file might not exist yet — that's fine, just skip
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	offset := state.offsets[path]

	// ── rotation detection: file is smaller than our cursor ─────
	if info.Size() < offset {
		offset = 0
	}

	if _, err := f.Seek(offset, 0); err != nil {
		return fmt.Errorf("seek %s: %w", path, err)
	}

	scanner := bufio.NewScanner(f)
	// raise buffer for very long NASL trace lines
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		fn(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan %s: %w", path, err)
	}

	// persist new cursor
	newOffset, _ := f.Seek(0, 1) // current position
	state.offsets[path] = newOffset
	return nil
}
