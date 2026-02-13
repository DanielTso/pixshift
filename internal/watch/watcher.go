package watch

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/DanielTso/pixshift/internal/codec"
	"github.com/DanielTso/pixshift/internal/pipeline"
	"github.com/fsnotify/fsnotify"
)

// Watcher monitors directories for new image files and converts them.
type Watcher struct {
	Pipeline     *pipeline.Pipeline
	OutputDir    string
	OutputFormat codec.Format
	Quality      int
	Recursive    bool
	Metadata     bool
	Verbose      bool
	OnConvert    func(pipeline.Result)

	// v0.5.0 fields
	DebounceDelay  time.Duration // configurable (default 500ms)
	IgnorePatterns []string      // glob patterns to skip (e.g. "*.tmp", ".git")
	MaxRetries     int           // retry on error (default 0)
	JobTemplate    pipeline.Job  // template with all transforms
}

// Watch starts watching the given directories for new/modified image files.
// It blocks until the context is cancelled.
func (w *Watcher) Watch(ctx context.Context, dirs []string) error {
	fsw, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}
	defer fsw.Close()

	for _, dir := range dirs {
		if w.Recursive {
			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() {
					return fsw.Add(path)
				}
				return nil
			})
			if err != nil {
				return fmt.Errorf("watch %s: %w", dir, err)
			}
		} else {
			if err := fsw.Add(dir); err != nil {
				return fmt.Errorf("watch %s: %w", dir, err)
			}
		}
	}

	delay := w.DebounceDelay
	if delay == 0 {
		delay = 500 * time.Millisecond
	}

	// Debounce map: track pending files to handle partial writes
	var mu sync.Mutex
	pending := make(map[string]*time.Timer)

	for {
		select {
		case <-ctx.Done():
			mu.Lock()
			for _, t := range pending {
				t.Stop()
			}
			mu.Unlock()
			return nil

		case event, ok := <-fsw.Events:
			if !ok {
				return nil
			}

			// Watch new directories when recursive
			if event.Op&fsnotify.Create != 0 && w.Recursive {
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					_ = fsw.Add(event.Name)
					if w.Verbose {
						fmt.Fprintf(os.Stderr, "watching new directory: %s\n", event.Name)
					}
					continue
				}
			}

			if event.Op&(fsnotify.Create|fsnotify.Write) == 0 {
				continue
			}

			path := event.Name
			ext := filepath.Ext(path)
			if !codec.IsSupportedExtension(ext) || w.shouldIgnore(path) {
				continue
			}

			mu.Lock()
			if t, exists := pending[path]; exists {
				t.Stop()
			}
			pending[path] = time.AfterFunc(delay, func() {
				mu.Lock()
				delete(pending, path)
				mu.Unlock()
				w.processFile(path)
			})
			mu.Unlock()

		case err, ok := <-fsw.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "watch error: %v\n", err)
		}
	}
}

// shouldIgnore checks if a file should be ignored based on the configured patterns.
func (w *Watcher) shouldIgnore(path string) bool {
	base := filepath.Base(path)
	for _, pattern := range w.IgnorePatterns {
		if matched, _ := filepath.Match(pattern, base); matched {
			return true
		}
	}
	return false
}

func (w *Watcher) processFile(path string) {
	maxAttempts := w.MaxRetries + 1
	if maxAttempts < 1 {
		maxAttempts = 1
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt-1)) * time.Second
			if backoff > 16*time.Second {
				backoff = 16 * time.Second
			}
			time.Sleep(backoff)
			if w.Verbose {
				fmt.Fprintf(os.Stderr, "retry %d/%d for %s\n", attempt+1, maxAttempts, path)
			}
		}

		err := w.processFileOnce(path)
		if err == nil {
			return
		}
		lastErr = err
		if w.Verbose {
			fmt.Fprintf(os.Stderr, "attempt %d/%d failed for %s: %v\n", attempt+1, maxAttempts, path, lastErr)
		}
	}

	if w.OnConvert != nil && lastErr != nil {
		w.OnConvert(pipeline.Result{
			Job:   pipeline.Job{InputPath: path},
			Error: lastErr,
		})
	}
}

func (w *Watcher) processFileOnce(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	inputFormat, err := codec.DetectFormat(f, path)
	f.Close()
	if err != nil {
		return err
	}

	// Don't convert if input matches output format
	if inputFormat == w.OutputFormat {
		return nil
	}

	outputPath := buildOutputPath(path, w.OutputDir, w.OutputFormat)

	// Start from JobTemplate to inherit all transforms
	job := w.JobTemplate
	job.InputPath = path
	job.OutputPath = outputPath
	job.InputFormat = inputFormat
	job.OutputFormat = w.OutputFormat
	if job.Quality == 0 {
		job.Quality = w.Quality
	}
	job.PreserveMetadata = w.Metadata

	inSize, outSize, err := w.Pipeline.Execute(job)
	if err != nil {
		return err
	}

	if w.OnConvert != nil {
		w.OnConvert(pipeline.Result{Job: job, InputSize: inSize, OutputSize: outSize})
	}
	return nil
}

func buildOutputPath(inputPath, outputDir string, format codec.Format) string {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	name := base[:len(base)-len(ext)] + codec.DefaultExtension(format)

	if outputDir != "" {
		return filepath.Join(outputDir, name)
	}
	return filepath.Join(filepath.Dir(inputPath), name)
}
