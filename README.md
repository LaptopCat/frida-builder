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
`alias frida-builder="/path/to/frida-builder/main"`

# Terminology
- **workflow** - an ordered list of presets, which describes the bundling steps
- **preset** - an action to perform along with options for it (usage of a utility, such as esbuild, terser or quickjs; or reuse of a different workflow)

# Usage
The commands below assume you have an alias set up.

## Run workflow
There are 3 default workflows:
- minimal:

A minimal workflow. Just bundles everything without optimizations using esbuild.

- optimized:

Bundles everything and does some optimizations on the code, such as tree-shaking, minification, mangling using esbuild and terser

- optimized-bytecode:

Same as optimized, but it also compiles the result to QuickJS bytecode using qjsc

You can run a workflow like so:
```sh
frida-builder workflow <name> <entrypoint>
```

For example:
```sh
frida-builder workflow optimized-bytecode index.ts
```

This will output the result into a file named `bundle`. You can change it like so:
```sh
frida-builder -o <output-filename> workflow <name> <entrypoint>
```

## Writing your own workflows
Workflows are loaded from JSON files. The structure is like this:
```json
{
    "workflow-name--must-be-lowercase": [
        {
            "Util": "esbuild", // available utilities: esbuild, terser and qjsc
            "EsbuildOptions": { // refer to https://pkg.go.dev/github.com/evanw/esbuild@v0.24.0/pkgapi#BuildOptions
            // EsbuildOptions is required when Util is esbuild
                "Bundle": true
            }
        },

        {
            "Util": "terser",
            "Options": [ // command-line options, required when Util is terser
            // refer to https://terser.org/docs/cli-usage/ for terser util
                "--compress",
                "--mangle"
            ]
        }
    ],
    "another-workflow": [
        {"ReuseWorkflow": "workflow-name--must-be-lowercase"}, // you can run other workflows like so (you can use it anywhere, not only at the start)
        {"Util": "qjsc"} // qjsc util compiles JavaScript to QuickJS bytecode. There are no options
    ]
}
```

frida-builder tries to load workflows from `workflows.json`, both in the binary path, and current directory. You can also load workflows from certain files:
```sh
frida-builder --workflows flows1.json,flows2.json workflow <name> <entrypoint>
```