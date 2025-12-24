---
description: File system and shell utility commands
---

# File System & Shell Utilities

This workflow defines safe file system and shell utility commands.

## Directory Listing

// turbo
1. List directory tree:
```bash
tree <path>
```

// turbo
2. List directory contents (ls):
```bash
ls <path>
```

// turbo
3. List directory contents (dir):
```bash
dir <path>
```

## File Operations

// turbo
4. View file contents:
```bash
cat <file>
```

// turbo
5. Create directories:
```bash
mkdir -p <path>
```

// turbo
6. Move/rename files:
```bash
mv <source> <destination>
```

// turbo
7. Delete files (Windows):
```bash
del <file>
```

// turbo
8. Set file permissions:
```bash
chmod <mode> <file>
```

## Search & Filter

// turbo
9. Find files:
```bash
find <path> <options>
```

// turbo
10. Pattern matching (grep):
```bash
grep <pattern> <files>
```

// turbo
11. Pattern matching (findstr - Windows):
```bash
findstr <pattern> <files>
```

// turbo
12. Word/line count:
```bash
wc <options> <file>
```

// turbo
13. Locate executables:
```bash
where <command>
```
