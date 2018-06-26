# unlinker [![Build Status](https://travis-ci.org/tagomoris/unlinker.svg?branch=master)](https://travis-ci.org/tagomoris/unlinker)

Single binary tool to delete old (expired) files, especially for log files.

## Usage

```
./unlinker /path/to/config/dir
```

Configuration directory may contain 1 or more JSON files to specify how to determine deleted files.

## Configuration

This tool requires `rule` and `path`, and some other fields (rule specific).

|name|description|
|----|-----------|
|rule|Name of rule [age,timestamp,mtime]|
|path|Full path of files, must contain just one glob (`*`)|

### Rule: age

This rule determines deleted file using age of files. For this rule, paths are supposed as `/path/to/file/access_log.1.gz`. Glob should match with an integer of age.

For this rule, parameter `age` is required. Files with an age equals to specified one will be deleted.

```json
{
  "rule": "age",
  "path": "/path/to/file/access_log.*.gz",
  "age": 4
}
```

With the configuration above, `access_log.4.gz` and `access_log.5.gz` will be deleted, but `access_log.3.gz` will not be deleted.

### Rule: timestamp

This rule determines deleted files using timestamp string in path. Glob should match with timestamp string.

```json
{
  "rule": "timestamp",
  "path": "/path/to/file/access_log.*.gz",
  "format": "%Y%m%d_%H",
  "expire_sec": 86400
}
```

With the configuration above, files with timestamp of 1 day before (or older) will be deleted.

### Rule: mtime

This rule determines deleted files using filesystem level last-modified time (mtime).

```json
{
  "rule": "mtime",
  "path": "/path/to/file/access_log.*.gz",
  "expire_sec": 86400
}
```

## How To

Test: `go test -v`

Build: `go build`

Release: (set `GITHUB_TOKEN` to release binaries)
 - `gox -os="linux darwin" -arch="386 amd64" -output "build/{{.Dir}}_{{.OS}}_{{.Arch}}"`
 - `ghr $VERSION build/`

## Versions

- v0.1.1: Cleanup code/test
- v0.1.0: First version

## License

MIT License

## Copyright

Satoshi Tagomori [@tagomoris] (tagomoris at gmail.com)
