# Kothay

A CLI for organising files and folders with tags.

**Kothay** (pronounced _ko-thai_) takes its name from the Bengali word **কোথায়**, meaning _where?_

## Installation

You can download one of the prebuilt binaries for your platform [here](https://github.com/shsiddhant/kothay/releases).
The binary can be placed somewhere on your `PATH`.

## Usage

```shell
Usage: kothay <command> [flags]

A CLI for organising files and folders with tags.

Flags:
  -h, --help         Show context-sensitive help.
      --db=STRING    Path to the database file
  -v, --verbose      Enable verbose output
      --version      Show version number and quit

Commands:
  tag --tags=TAGS,... <paths> ... [flags]
    Tag files or directories.

  untag --tags=TAGS,... --all <paths> ... [flags]
    Remove tags from files or directories.

  find [<tags> ...] [flags]
    Find tagged files and directories.

  ls [<path>] [flags]
    List tags for a file/directory, or all tags.

  version [flags]
    Show detailed version information

Run "kothay <command> --help" for more information on a command.
```

### Tag files and directories

Attach one or more tags using `-t` / `--tags` (repeatable or comma-separated). Hierarchical tags separated by `/` are supported:

```shell
kothay tag -t Lang/Go -t Code ./my-project
kothay tag -t docs,reference README.md
```

### Find tagged resources

Search by full tag or tag segment, or omit arguments to list all tagged items:

```shell
# Matches files/directories tagged with "Lang/Go" or "Go"
kothay find Go

# Filter results by type
kothay find --files Code
kothay find --dirs Lang/Go

# List all tagged files and directories
kothay find
```

### List tags

List tags attached to a specific path, or list all known tags:

```shell
# List tags on a specific file or directory
kothay ls ./my-project

# List all tags in the database
kothay ls
```

### Remove tags

Remove specific tags or clear all tags from given paths:

```shell
kothay untag -t Code ./my-project
kothay untag --all ./my-project
```

## License

This project is licensed under the GNU General Public License v3.0.

See [LICENSE](LICENSE) for details.
