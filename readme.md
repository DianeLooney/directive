# Directive

Package for parsing DRY data.

## Example
```
name "something"
other_name "something else"
version "30"
[author "diane" "john" "anonymous"]

@note { freq "440"; duration "1.beat" }

measure {
    [note {} {} {} {}]
}
```

## EBNF
```
document           = { directive | repeated_directive }
directive          = [ "@" ], identifier, value, [ ";" ]
repeated_directive = "[", [ "@" ], identifier, values "]"

identifier         = /([a-zA-Z][a-zA-Z0-9_]+)/
values             = { value, [","] }
value              = object | string
object             = "{", { directive | repeated_directive }, "}"
string             = /"((?:[^"\\]|\\.)*)"/
                   | /'((?:[^'\\]|\\.)*)'/
                   | /`([^`]*)`/
```
