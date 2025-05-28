//go:build linux || darwin
// +build linux darwin

package collector

import (
	"context"
	"revnoa/utils"
	"sync"
	"time"

	"github.com/hpcloud/tail"
)

type UnixTailer struct {
	files       []string
	buffer      []string
	bufferCount int
	flushTicker *time.Ticker
	sendFunc    func([]string)

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewTailer(files []string, bufferCount int, flushSec int, sendFunc func([]string)) Tailer {
	return NewUnixTailer(files, bufferCount, flushSec, sendFunc)
}

func NewUnixTailer(files []string, bufferCount int, flushSec int, sendFunc func([]string)) *UnixTailer {
	ctx, cancel := context.WithCancel(context.Background())
	return &UnixTailer{
		files:       files,
		bufferCount: bufferCount,
		flushTicker: time.NewTicker(time.Duration(flushSec) * time.Second),
		sendFunc:    sendFunc,
		ctx:         ctx,
		cancel:      cancel,
	}
}

func (t *UnixTailer) Start() error {
	for _, file := range t.files {
		file := file
		t.wg.Add(1)
		go func() {
			defer t.wg.Done()

			tailer, err := tail.TailFile(file, tail.Config{
				Follow:    true,
				ReOpen:    true,
				MustExist: false,
			})
			if err != nil {
				utils.ErrorLogger.Printf("Failed to tail file: %s, err: %v", file, err)
				return
			}
			defer tailer.Cleanup()

			for {
				select {
				case <-t.ctx.Done():
					utils.InfoLogger.Printf("Stopped tailing: %s", file)
					return
				case line, ok := <-tailer.Lines:
					if !ok {
						utils.WarnLogger.Printf("Tail stopped for file: %s", file)
						return
					}
					if line.Err != nil {
						utils.WarnLogger.Printf("Tail error on %s: %v", file, line.Err)
						continue
					}
					t.buffer = append(t.buffer, line.Text)
					if len(t.buffer) >= t.bufferCount {
						t.flush()
					}
				}
			}
		}()
	}

	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		for {
			select {
			case <-t.ctx.Done():
				return
			case <-t.flushTicker.C:
				if len(t.buffer) > 0 {
					t.flush()
				}
			}
		}
	}()

	return nil
}

func (t *UnixTailer) Stop() {
	t.cancel()
	t.flushTicker.Stop()
	t.wg.Wait()
}

func (t *UnixTailer) flush() {
	t.sendFunc(t.buffer)
	t.buffer = nil
}
