# xos

The standard library's `os`, with one difference on Windows: every file `xos`
opens or creates carries `FILE_SHARE_DELETE`, so another process can rename or
delete the file while this one holds it, as it can on every other platform.
`Rename` replaces a file readers hold open, through the POSIX-semantics rename
where the volume has it. Everything else is `os`, call for call: the same
names, the same errors, checked against `os` on the same filesystem state on
every platform.

```go
import "github.com/forkcloser/xos"

f, err := xos.OpenFile(path, os.O_RDWR|os.O_CREATE, 0o644)
err = xos.WriteFile(path, data, 0o644) // writes beside, then renames over
```

A program that opens, creates, renames or reads files that another process may
replace under it uses `xos` where it would use `os`. The one thing `os` does on
Windows that `xos` does not is the `\\?\` prefix for a long path on Windows 10
before 1703 and Server 2016; later releases make the process long-path aware
and neither needs it.

Moved from primordium's `filesystem/xos`, which it replaces.

## Developing

Every workflow runs through `just`: `just lint`, `just test`, `just fix`;
`just --list` shows the rest. The contributor rules are in `AGENTS.md`.
