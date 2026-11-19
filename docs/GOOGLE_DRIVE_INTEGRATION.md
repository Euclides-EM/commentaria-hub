# Overview

In general, we use Google Drive for two purposes:

1. As a staging area for facsimile PDFs and diagram crops before they are imported into the hub server.
2. As a backup storage for facsimile PDFs, which can be downloaded through the API when needed.

This document describes how to upload facsimile PDFs and diagram crops to the hub server using a Google Drive inbox folder. The server imports those files, sets them in the local filesystem, and updates the facsimile metadata accordingly.

# Initial setup

On your local machine, set up rclone with a new remote for your Google Drive account:

```bash
sudo -v ; curl https://rclone.org/install.sh | sudo bash
rclone config // an interactive command, choose the following:
n // new remote
G // remote name
19 // GDrive
4***k.apps.googleusercontent.com // clientid
G***9G // client secret
1 // full access
Service account file leave empty
No // dont edit advanced config
Yes // authenticate with browser for the your real account, hit the “I trust Liri” warnings…
```

Then, to get the config file path, run:

```bash
rclone config file
```

Copy the contents of that file.

On the server, install rclone as root:

```bash
sudo -v ; curl https://rclone.org/install.sh | sudo bash
```

Then switch to the `euclides` user and create the config file with the same contents as your local machine:

```bash
sudo -iu euclides
rclone config file
```

This will show you the path where rclone expects the config file, likely `~/.config/rclone/rclone.conf`. Create that file and paste the contents from your local machine.

Now, you can follow the instructions in [Hub deployment](./HUB_DEPLOYMENT.md) to set the relevant environment variables for the facsimile PDF directory and the Google Drive inbox folder ID.

# Day to day usage

## Upload diagram crops

If diagram crops are produced elsewhere, for example on the GPU farm, upload them to the same Google Drive inbox used for facsimile PDFs.

The archive contents should use this layout:

```text
<edition_key>/crops/*.jpg
```

That edition-key layout is still supported for unambiguous existing data. If an edition has multiple local facsimiles and the crops belong to one specific scan, prefer a facsimile-specific key that matches the PDF basename/facsimile name:

```text
<edition_key>_bnf/crops/*.jpg
<edition_key>_google/crops/*.jpg
```

For multi-volume editions, keep one crop directory per volume:

```text
<edition_key>_vol1/crops/*.jpg
<edition_key>_vol2/crops/*.jpg
```

On the GPU farm, package one or more finished crop directories as `.zip`, `.tar.gz`, or `.tgz`. The directory names inside the archive must be the final edition, volume, or facsimile-specific keys:

```bash
cd /path/to/gpu-output
tar -czf /tmp/commentaria-diagram-crops.tar.gz Venice_1482
```

For several editions or volumes:

```bash
cd /path/to/gpu-output
tar -czf /tmp/commentaria-diagram-crops.tar.gz Venice_1482 Paris_1615_vol1 Paris_1615_vol2
```

Upload that archive to the facsimile Google Drive inbox folder. The server import accepts both PDFs and crop archives from that folder.

Then click **Import Facsimiles** from the user menu in the hub app. The import endpoint downloads the archive from Drive, installs the crop directories under `/data/euclides/commentaria-hub/facsimiles/diagrams`, clears stale metadata for the affected keys, regenerates diagram metadata, and deletes the successfully imported archive from Drive.

Check one uploaded crop through the backend and through nginx:

```bash
curl -I http://127.0.0.1:8090/facsimiles/diagrams/Venice_1482/crops/5_Content_Illustration_4.jpg || true
curl -I https://euclides.huma-num.fr/commentaria/facsimiles/diagrams/Venice_1482/crops/5_Content_Illustration_4.jpg || true
```

## Import or download facsimile PDFs

The easiest day-to-day path is a Google Drive inbox folder. Upload one or more PDFs or diagram crop archives to that folder. A single PDF uses the `<edition_key>.pdf` naming convention. To keep multiple PDFs for one edition, append a distinguishing label after the edition key, for example `<edition_key>_bnf.pdf`, `<edition_key>_google.pdf`, or `<edition_key>_vol1_bnf.pdf`. Every matching file becomes a separate facsimile row linked to `<edition_key>`. Unknown PDF prefixes are skipped instead of creating facsimiles for non-existent editions. Diagram crop archives must contain crop directories as described above. Then click **Import Facsimiles** from the user menu in the hub app, or call the API endpoint below.

The import endpoint:

- lists PDFs and crop archives in `FACSIMILES_GDRIVE_FOLDER_ID` using `rclone`;
- copies them into `FACSIMILES_PDF_DIR`;
- creates or updates the local facsimile DB rows;
- installs crop archives into `FACSIMILES_DIAGRAMS_PATH`;
- regenerates diagram crop metadata;
- deletes only the successfully imported files from the Drive folder.

To call it manually:

```bash
curl -X POST \
  -H "Authorization: Bearer <github-token>" \
  http://127.0.0.1:8090/api/v1/facsimilies/import-from-drive
```

The endpoint returns JSON like:

```json
{
  "importedPdfs": ["Venice_1482.pdf", "Paris_1615.pdf"],
  "importedDiagramArchives": ["commentaria-diagram-crops.tar.gz"],
  "importedDiagramCrops": ["Venice_1482", "Paris_1615_vol1", "Paris_1615_vol2"],
  "skipped": [],
  "deleted": [
    "Venice_1482.pdf",
    "Paris_1615.pdf",
    "commentaria-diagram-crops.tar.gz"
  ]
}
```

To download a stored PDF through the API, pass the facsimile ID and an auth bearer token:

```bash
curl -fL \
  -H "Authorization: Bearer <github-token>" \
  -o Venice_1482.pdf \
  http://127.0.0.1:8090/api/v1/facsimilies/<facsimile-id>/pdf
```

To delete a facsimile, use the authenticated delete endpoint:

```bash
curl -f -X DELETE \
  -H "Authorization: Bearer <github-token>" \
  http://127.0.0.1:8090/api/v1/facsimilies/<facsimile-id>
```

The endpoint deletes both the database record and a PDF managed under `FACSIMILES_PDF_DIR`, so the startup scan does not import it again. It refuses to delete a facsimile referenced by a dataset; delete those datasets first. PDFs outside the configured server-managed directory are not removed.

After importing PDFs, connect each facsimile to its shelfmark from the edition edit page. A single edition can have multiple facsimiles, but each shelfmark can be connected to at most one facsimile. For bulk cleanup, use the temporary facsimile mapping ZIP from the edition edit page: use `shelfmarks.csv` as the lookup table, edit only `shelfmark_id` and `facsimile_connection_confirmation_status` in `facsimiles.csv`, then upload `facsimiles.csv`.
