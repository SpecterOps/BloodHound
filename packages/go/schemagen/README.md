# CUE Lang Driven Code Generation

## TODO

- Validation and testing.
- More documentation.
- Fold in generation of OpenAPI types.

## About

This command will generate graph schema files for Go and TypeScript. It currently
directly depends on the Cue module found in `packages/cue/bh` to generate its schema.

## Folders

### generator

This directory contains the main library for generating for each of our supported languages

### tsgen

`tsgen` is a code-generation framework for [TypeScript](https://www.typescriptlang.org) written in golang. It is loosely
inspired in both pattern and structure by [Jennifer](https://github.com/dave/jennifer) a code generator for golang
written in golang.

`tsgen` is not complete and only implements the bare minimum sub-set of TypeScript syntax constructs required to produce
schema types and helper functions. General TypeScript code compatibility is not a goal of `tsgen`.

## Generating The Latest Schema

```bash
just schemagen
```
