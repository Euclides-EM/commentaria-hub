# Facsimile Media Backup

This backup is for the large media files that should not be bundled into the app ZIP backup:

- facsimile PDFs: `/data/euclides/commentaria-hub/facsimiles/pdfs`
- diagram crops: `/data/euclides/commentaria-hub/facsimiles/diagrams`

Use the existing hub backup for the database, models, and metadata. Use this media backup for the heavy PDF and crop directories.

The process uses `rclone` so it can copy from the server to either:

- a physical drive mounted on your computer;
- a cloud target such as Google Drive, S3, Backblaze B2, or another SFTP server.

`rclone copy` checks the destination before copying. Files that already match are skipped, so rerunning the process is normal and safe. No large temporary file is created on your computer; data streams from the server to the destination. For local drives, partial files are written on the destination drive and corrected on the next run.

## One-Time Setup

Install `rclone` on your computer:

```bash
brew install rclone
```

Create an SFTP remote for the hub server:

```bash
rclone config
```

Recommended values:

```text
name: commentaria-server
Storage: sftp
Host: euclides.huma-num.fr
User: euclides
```

Use SSH key authentication if possible. Test that `rclone` can see the media root:

```bash
rclone lsd commentaria-server:/data/euclides/commentaria-hub/facsimiles
```

You should see at least:

```text
pdfs
diagrams
```

## Physical Drive Backup

Connect the drive and make sure it is mounted. On macOS this will usually be under `/Volumes`, for example:

```text
/Volumes/CommentariaBackup
```

Run a dry run first:

```bash
./scripts/backup_facsimile_media.sh \
  --source commentaria-server:/data/euclides/commentaria-hub/facsimiles \
  --dest /Volumes/CommentariaBackup/commentaria-hub/facsimiles \
  --dry-run
```

Then run the real backup:

```bash
./scripts/backup_facsimile_media.sh \
  --source commentaria-server:/data/euclides/commentaria-hub/facsimiles \
  --dest /Volumes/CommentariaBackup/commentaria-hub/facsimiles
```

Logs are written on the destination drive by default:

```text
/Volumes/CommentariaBackup/commentaria-hub/facsimiles/_backup_logs/
```

The important log files are:

- `*.summary.txt`: command settings and high-level output;
- `*.log`: detailed `rclone` log;
- `*.combined.log`: per-file result list. Lines indicate copied, skipped, changed, deleted, or failed files.

If the process stops, reconnect the drive and run the same command again. Matching files will be skipped, partial or changed files will be copied again, and the log will show what happened in the new run.

## Cloud Backup

Configure the cloud destination with `rclone config`, then use the same script with a remote destination.

For example, if your Google Drive remote is named `gdrive`:

```bash
./scripts/backup_facsimile_media.sh \
  --source commentaria-server:/data/euclides/commentaria-hub/facsimiles \
  --dest gdrive:commentaria-hub-media-backup/facsimiles
```

For cloud destinations, logs default to:

```text
./logs/facsimile-media-backup/
```

You can put logs on the physical drive or anywhere else:

```bash
./scripts/backup_facsimile_media.sh \
  --source commentaria-server:/data/euclides/commentaria-hub/facsimiles \
  --dest gdrive:commentaria-hub-media-backup/facsimiles \
  --log-dir /Volumes/CommentariaBackup/commentaria-hub/logs/cloud-media-backup
```

## Optional Mirror Mode

By default, the script only copies new or changed files. It does not delete files from the backup if they were removed on the server.

To make the backup an exact mirror, add `--delete-extra`:

```bash
./scripts/backup_facsimile_media.sh \
  --source commentaria-server:/data/euclides/commentaria-hub/facsimiles \
  --dest /Volumes/CommentariaBackup/commentaria-hub/facsimiles \
  --delete-extra
```

Use this only when you are confident that server deletions should also remove files from the backup.

## Verification

Count the server files:

```bash
rclone size commentaria-server:/data/euclides/commentaria-hub/facsimiles/pdfs
rclone size commentaria-server:/data/euclides/commentaria-hub/facsimiles/diagrams
```

Count the backup files:

```bash
rclone size /Volumes/CommentariaBackup/commentaria-hub/facsimiles/pdfs
rclone size /Volumes/CommentariaBackup/commentaria-hub/facsimiles/diagrams
```

For a stronger check, run:

```bash
rclone check \
  commentaria-server:/data/euclides/commentaria-hub/facsimiles \
  /Volumes/CommentariaBackup/commentaria-hub/facsimiles \
  --one-way \
  --size-only
```

Use the cloud destination path instead of the `/Volumes/...` path when checking a cloud backup.

## Recommended Routine

1. Connect the physical drive.
2. Run the dry run.
3. Run the real backup.
4. Check the newest `_backup_logs/*.summary.txt` and `_backup_logs/*.combined.log`.
5. Run `rclone size` or `rclone check` when you want extra confidence.
6. Eject the drive.

Running this from time to time is enough. The process is designed to be boring: same command, repeatable logs, and no large temporary files on the computer.
