# frida-builder
CLI for bundling your frida scripts to javascript or quickjs bytecode

# Installing
1. Install [Go](https://go.dev/doc/install)
2. `git clone https://github.com/laptopcat/frida-builder`
3. `cd frida-builder`
4. `go get`
5. `go build main.go`
6. Install Terser (`npm i terser -g`)
7. Install QuickJS (`git clone https://github.com/frida/quickjs; cd quickjs; make install`)