# golang_git
Implementation of git internal commands in Go language.

## Support commands
* `gogit init <path>`
* `gogit hash-object <file_name>`
* `gogit cat-file -t|-s|-p <object_hash>`
* `gogit ls-tree <tree_hash>`
* `gogit mktree`
* `gogit checkout [-path <path>] <tree_hash>`
