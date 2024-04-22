<script setup>
  import {data} from './uploading.data.js'
</script>

# Uploading with exporters

ortfo/db comes with built-in exporters that cover most use cases for uploading the generated database.json file somwhere:

## Cloud service providers

ortfo/db supports uploading to many cloud service providers by leveraging [rclone](https://rclone.org/).

This allows us to support a wide range of cloud services, such as Google Drive, Dropbox, and many others.

::: details See all supported services

This list is automatically generated from [rclone's documentation](https://rclone.org/overview#features).

<ul>
  <li v-for="service in data.backends">{{ service }}</li>
</ul>
:::

### Setup

::: info
In the future, ortfo/db might support configuring the remotes by itself
:::

1. Install rclone from [rclone.org](https://rclone.org/downloads/)
2. [Configure a remote](https://rclone.org/docs/#configure) with `rclone config`
3. Keep in mind the name you gave to the remote, you'll need it later

### Configuration

```yaml [ortfodb.yaml]
exporters:
  cloud:
    # Path to the folder where you want to upload the database.json file
    path: projects/
    # Filename to copy in the folder as (no renaming by default)
    name: my_database.json
    # Name of the rclone remote(s)
    # This is a list because you can tell ortfo/db to upload to multiple rclone remotes in one go.
    remotes: [my_remote]
```

## Web servers

### SSH

You can upload the database file to a server via SSH using the `ssh` exporter.

#### Configuration

```yaml [ortfodb.yaml]
exporters:
  ssh:
    ssh: user@host:/path/to/database.json
    rsync: true # true to use rsync, false to use scp
```

### (S)FTP

The `ftp` exporter allows you to upload the generated database.json file to a remote server using the FTP protocol (or SFTP) protocol.

#### Configuration

```yaml [ortfodb.yaml]
exporters:
  ftp:
    username: your username
    password: your password
    host: the host of the server
    path: the path to where the file should be uploaded
    secure: true # true to use SFTP, false to use FTP
```

## Git repositories

Use the `git` exporter

### Configuration

```yaml [ortfodb.yaml]
exporters:
  git:
    # URL to the Git repository
    url:

    # Commit message used
    commit:

    # Where to copy the database.json file in the repository
    path: database.json

    # Additional flags to pass to git commit
    git_commit_flags: ""

    # Additional flags to pass to git push
    git_push_flags: ""

    # Additional flags to pass to git clone
    git_clone_flags: ""

    # Additional flags to pass to git add
    git_add_flags: ""
