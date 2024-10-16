# Contributing

If you would like to contribute code to the nbde-tang-server project, you can do so
through GitHub by forking the repository and sending a Pull Request.

When submitting Golang code, please make every effort to follow existing conventions
and style in order to keep the code as readable as possible. Please also make sure
all tests pass by running `go test`, and format your code with `go fmt`.
We also recommend using `golint` and `staticcheck` for static code analysis,
as well as `gosec` for security static code inspection.

On the other hand, when submitting Bash code, please verify also that your scripts
are compliant to Bash by using `shellcheck` command on them.

Last, but not least, in case your Pull Request contains Markdown documentation files,
ensure no typographic or orthographic issues are included.

Configured pipelines will run part of the previous checks, but running them before pushing
will contribute to avoid using unnecessary CI resources.

And, the most important thing ... Thanks for your contribution!!!
