# NIT
---
A minimal Git implementation written in go for learning purposes only.

## Goal
- To deepen my Golang understanding
- To understand Git's internal architecture by building a functional clone, including:
	- Content-addressable object database
	- Index (staging area)
	- Trees, commits, and references
	- Basic CLI commands

## Features implemented
### Commands

| Command | Description                                                |
| ------- | ---------------------------------------------------------- |
| init    | initialize a new repository<br>                            |
| add     | add files to the index (stage files)                       |
| status  | show the working tree status (staged, modified, untracked) |
| commit  | record changes to the repository                           |
| log     | show commit logs with pager support                        |
| diff    | show changes between commit and working tree               |
| help    | show help message                                          |

### Internals
- Loose object storage (zlib compression + SHA-1)
- Proper Git object format compatibililty
- Tree building from index (recursive)
- Index read/write
- Colorized output + pager support 

## Quick start
### 1. Build
```bash
go build -o=./bin ./...
```
### 2. Usage
```bash
# Initialize a repository
./bin/nit init

# Stage files
./bin/nit add <files...>

# Commit
./bin/nit commit -m "message"

# Check status
./bin/nit status
./bin/nit status --porcelain

# View history
./bin/nit log

# View help message
./bin/nit help

# View changes between latest commit and working tree
./bin/nit diff
```

## Future Plans
- Branching and merging support
- Performance optimization

## Technologies
- Go (1.25+)
- Standard library only (no dependency)

## Note 
- I only tested this code on MacOs
