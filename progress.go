package main

import (
	"fmt"
	"os"
	"time"

	"github.com/dustin/go-humanize"
	"go.uber.org/zap"
)

func NewProgressPrinter(logger *zap.Logger, srcSize uint64, destName string) *ProgressPrinter {

	return &ProgressPrinter{
		srcSizeHuman: humanize.Bytes(srcSize),
		logger:       logger,
		ticker:       nil,
		destName:     destName,
	}
}

// Logs bytes downloaded every 5 seconds
type ProgressPrinter struct {
	srcSizeHuman string
	logger       *zap.Logger
	stop         chan struct{}
	ticker       *time.Ticker
	destName     string
}

func (pp *ProgressPrinter) Start() {

	pp.stop = make(chan struct{})

	ticker := time.NewTicker(time.Second * 5)
	go func(destName string, srcSizeHuman string, ticker *time.Ticker, stop chan struct{}) {

		for {
			select {
			case <-ticker.C:
				fi, err := os.Stat(destName)
				if err != nil {
					pp.logger.Error("Failed to stat "+destName, zap.Error(err))
				} else {
					pp.logger.Info(fmt.Sprintf("Downloading to %s loaded %s / %s", destName, humanize.Bytes(uint64(fi.Size())), srcSizeHuman))
				}
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}(pp.destName, pp.srcSizeHuman, ticker, pp.stop)
}

func (pp *ProgressPrinter) Stop() {
	pp.stop <- struct{}{}
}
