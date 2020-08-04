# annolog-go

## Usage

annolog [-build "go build flags"] packagespec < input > output

## Description

Consume a log file generated by a Go program using 
[github.com/hashicorp/go-hclog](https://github.com/hashicorp/go-hclog),
output the same lines but with annotations appended describing the
code that generated each line.

In principle there's very little tying this specifically to hclog, but
for now that's the only logging library supported.

## Examples

These examples all use Vault but any codebase using hclog should work.

Example invocation (from a Vault OSS checkout):

```
annolog -build "-mod=vendor" github.com/hashicorp/vault/... 
```

Example invocation (from a Vault enterprise checkout):

```
annolog -build "-tags=enterprise -mod=vendor" github.com/hashicorp/vault/... 
```

Example input:
```
2020-01-21T15:49:20.959-0500 [DEBUG] perf-sec.core2.expiration: collecting leases
```

Example output:
```
2020-01-21T15:49:20.959-0500 [DEBUG] perf-sec.core2.expiration: collecting leases /home/ncc/gh/vault-enterprise/vault/expiration.go:464:17 (Restore)
```

## Notes

annolog errs on the side of simplicity rather than aiming for perfection.
This manifests both in the rudimentary log parsing code and in the source
code analysis.  Despite being simplistic it works surprisingly well, and 
is pretty speedy thanks to the Go compiler being so fast.

### Parsing

* lines not starting with "202" are ignored; this is based on the assumption
  that every log line will start with a date in YYYY-something format, and is
  an optimization to avoid trying to deal with multi-line log entries
* everything on a line up to and including the first `]` is ignored
* the remainder is assumed to be of the form "message: k1=v1, k2=v2, ...";
  only `message` is searched for in the Go source tree
* colons in the message or KVs may result in no annotation being emitted

### Source lookup

* annolog looks up messages by looking at the first argument in calls to methods 
  named `Trace`, `Debug`, `Info`, `Warn`, or `Error`; no attempt is made to
  validate that the method receiver is actually an hclog Logger
* only handles literal values in the message; use of a constant, variable,
  or other expression to generate the message will prevent annotation

# TODO

Support JSON formatted logs.

