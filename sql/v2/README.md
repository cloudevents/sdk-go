# CloudEvents Expression Language Go implementation

## Development guide

To regenerate the parser, make sure you have [ANTLR4 installed](https://github.com/antlr/antlr4/blob/master/doc/getting-started.md) and then run:

```shell
antlr4 -Dlanguage=Go -package gen -o gen -visitor -no-listener CESQLParser.g4
```
