package main

import (
	"crypto/sha1"
	"fmt"
	"os"
	"strings"

	"github.com/dgraph-io/badger/v2"
	"github.com/pkg/sftp"
)

// Find new data to download
// Will not walk sub-directories to root, and will consider a sub-directory to be a "file" rather then a collection of
// files
func FindSingleLevelData(client *sftp.Client, root string) ([]os.FileInfo, error) {

	return client.ReadDir(root)
}


// Returns a boolean if the file should be downloaded
func CheckShouldDownload(txn *badger.Txn, root string, file os.FileInfo) (bool, error) {

	_, err := txn.Get(GenerateBadgerKey(file.Name(), root))
	if err != nil {
		if err == badger.ErrKeyNotFound {
			return true, nil
		}
	}

	return false, err
}

// Mark a file as downloaded
func MarkFileDownloaded(txn *badger.Txn, root string, file os.FileInfo) error {
	return txn.Set(GenerateBadgerKey(file.Name(), root), GenerateBadgerKey(file.Name(), root))
}

// Generate a key for a file name / path
func GenerateBadgerKey(file, root string) []byte {

	file = strings.ToLower(strings.Trim(file, "/"))
	root = strings.ToLower(strings.Trim(root, "/"))
	return []byte(fmt.Sprintf("% x", sha1.Sum([]byte("/"+root+"/"+file))))

}