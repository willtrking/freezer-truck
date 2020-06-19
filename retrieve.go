package main

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/sftp"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh"
)

type RetrievalOptions struct {
	// Remote path where files are stored
	RemoteFileRoot string

	// Local path where files are moved to after being downloaded
	// Files will be complete when moved here
	// MUST be on the same hard drive as TempFileRoot
	LocalFileRoot string

	// Local path where files are downloaded to
	// Contains in-progress downloads
	// MUST be on the same hard drive as LocalFileRoot
	TempFileRoot string

	SshClient *ssh.Client

	BadgerDb *badger.DB

	ConcurrencyLimiter *Limiter
}

func RetrieveNewFiles(logger *zap.Logger, opts RetrievalOptions) error {

	sftpClient, err := sftp.NewClient(opts.SshClient)
	if err != nil {
		return err
	}

	_, err = MkdirIfNotExists(opts.LocalFileRoot)
	if err != nil {
		return err
	}

	_, err = MkdirIfNotExists(opts.TempFileRoot)
	if err != nil {
		return err
	}

	// Locate data to download
	files, err := FindSingleLevelData(sftpClient, opts.RemoteFileRoot)
	if err != nil {
		return err
	}

	for _, toDownload := range files {

		opts.ConcurrencyLimiter.Acquire()

		go func(file os.FileInfo) {
			defer opts.ConcurrencyLimiter.Release()
			err := retrieveFile(logger, sftpClient, file, opts)
			if err != nil {
				logger.Error("Failed to download file "+file.Name()+", will be retried during next batch", zap.Error(err))
			}
		}(toDownload)
	}

	return nil

}

// Check to see if a file has not yet been downloaded, and download it to tmp dir if it is
// Update DB after to mark as downloaded, and move out of tmp folder.
func retrieveFile(logger *zap.Logger, sftpClient *sftp.Client, file os.FileInfo, opts RetrievalOptions) error {

	return opts.BadgerDb.Update(func(txn *badger.Txn) error {

		shouldDownload, err := CheckShouldDownload(txn, opts.RemoteFileRoot, file)

		if err != nil {
			return err
		}
		if !shouldDownload {
			logger.Debug("Skipped remote file " + CreateFilePath(opts.RemoteFileRoot, file))
			return nil
		}

		if file.IsDir() {

			dirWalker := sftpClient.Walk(CreateFilePath(opts.RemoteFileRoot, file))

			for dirWalker.Step() {
				if err := dirWalker.Err(); err != nil {
					return err
				}

				dirLocalPath := filepath.Join(opts.TempFileRoot, strings.Replace(filepath.Dir(dirWalker.Path()), opts.RemoteFileRoot, "", 1))
				didCreate, err := MkdirIfNotExists(dirLocalPath)
				if err != nil {
					return err
				}

				if didCreate {
					logger.Info("Created local dir " + dirLocalPath)
				}

				stepStat := dirWalker.Stat()
				if stepStat.IsDir() {
					continue
				}

				err = DownloadFile(logger, sftpClient, dirWalker.Path(), CreateFilePath(dirLocalPath, stepStat))
				if err != nil {
					return err
				}
			}

		} else {
			err = DownloadFile(logger, sftpClient, CreateFilePath(opts.RemoteFileRoot, file), CreateFilePath(opts.TempFileRoot, file))
		}

		if err != nil {
			return err
		}

		err = MarkFileDownloaded(txn, opts.RemoteFileRoot, file)
		if err != nil {
			return err
		}

		return nil

	})
}

// Create a file path for a root + file
func CreateFilePath(root string, file os.FileInfo) string {

	fileName := strings.Trim(file.Name(), "/")
	root = strings.Trim(root, "/")

	return "/" + root + "/" + fileName
}
