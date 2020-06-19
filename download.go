package main

import (
	"io"
	"os"

	"github.com/pkg/sftp"
	"go.uber.org/zap"
)

// Download a file, and emit progress information
func DownloadFile(logger *zap.Logger, client *sftp.Client, src string, dest string) error {

	logger.Info("Starting download of " + src)
	destFile, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	if err != nil {
		return err
	}

	defer destFile.Close()

	srcFile, err := client.Open(src)

	if err != nil {
		return err
	}

	defer srcFile.Close()

	statData, err := srcFile.Stat()
	if err != nil {
		return err
	}

	printer := NewProgressPrinter(logger, uint64(statData.Size()), dest)
	printer.Start()

	if _, err = io.Copy(destFile, srcFile); err != nil {
		printer.Stop()
		return err
	}

	printer.Stop()

	return destFile.Sync()
}
