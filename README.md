# frida-builder
CLI for bundling your frida scripts to javascript or quickjs bytecode

# Installing
1. Install [Git](https://git-scm.com/)
2. `git clone https://github.com/laptopcat/frida-builder`
3. `cd frida-builder`
4. Install [Go](https://go.dev/doc/install)
5. `go get`
6. `go build main.go`
7. Install Terser (`npm i terser -g`)
8. Install QuickJS (`git clone https://github.com/frida/quickjs; cd quickjs; sudo make install`)

You can also add an alias to your shell profile for convenience:
`alias frida-builder="/path/to/built/binary"`

# Terminology
- **workflow** - an ordered list of presets, which describes the bundling steps
- **preset** - an action to perform along with options for it (usage of a utility, such as esbuild, terser or quickjs; or reuse of a different workflow)

# Usage
The commands below assume you have an alias set up.

## Run workflow
```sh
frida-builder workflow <name> <entrypoint>
```

For example:
```sh
frida-builder workflow optimized-bytecode index.ts
```

This will output the result into a file named `bundle`. You can change it like so:
```sh
frida-builder -o <output-path> workflow <name> <entrypoint>
```

## Writing your own workflows
