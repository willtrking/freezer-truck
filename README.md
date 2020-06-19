freezer-truck
===============

`freezer-truck` allows users to synchronize files on a cron schedule from a remote server over SSH. It was primarily 
created to download hot copies of files to a staging area before they are tagged and sorted for cold storage. It 
downloads files over SSH with SFTP and provides the ability to concurrently do so.


# Why not XXX existing solution?

Most existing solutions use the contents of local and remote folders to determine if synchronization is necessary. Since
`freezer-truck` was designed with the expectation that another process would move files out of the directory it 
downloads to, it instead uses a local DB ([badger](https://github.com/dgraph-io/badger)) to keep track of which files 
have been loaded. This prevents the synchronization process from re-downloading files that have already been loaded, but simply
moved.


# Configuration

Configuration is done through the following environment variables.

- `TRUCK_SCHEDULECRON` **Required** Crontab-format schedule that `freezer-truck` should follow to download files.
- `TRUCK_HOST` **Required** SSH host to load files from
- `TRUCK_USERNAME` **Required** Username to use when connecting to host over SSH.
- `TRUCK_PRIVATEKEY` **Optional if TRUCK_PASSWORD is specified** SSH privatekey to use.
- `TRUCK_PASSWORD` **Optional if TRUCK_PRIVATEKEY is specified** Either the user's login password for SSH access, OR
the password to the specified private key file if `TRUCK_PRIVATEKEY` is also set
- `TRUCK_INSECUREKNOWNHOSTS` **Optional, default false** Ignore known hosts file. **NOT RECOMMENDED FOR PRODUCTION**
- `TRUCK_BADGERPATH` **Required** Path where [badger](https://github.com/dgraph-io/badger) database should be created.
- `TRUCK_CONCURRENCYLIMIT` **Required** How many files should be concurrently downloaded
- `TRUCK_REMOTEFILEROOT` **Required** Remote folder path where files should be loaded from
- `TRUCK_LOCALFILEROOT` **Required** Local folder path where complete downloads should be placed. In-progress downloads
will be located under the path specified by `TRUCK_TEMPFILEROOT`
- `TRUCK_TEMPFILEROOT` **Required** Local folder path where in-progress downloads should be placed. After completion,
files will be moved to the path specified by `TRUCK_LOCALFILEROOT`
