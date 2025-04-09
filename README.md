# godoc-mcp-server

search golang packages and their docs from pkg.go.dev, provide the infomation to llm as mcp server

## Install

You can download the compiled binary from the [releases page](https://github.com/yikakia/godoc-mcp-server/releases/tag/latest).

Or you can install it from source, compile it by yourself:

```shell
go install github.com/yikakia/godoc-mcp-server/cmd/godoc-mcp-server@latest
```

## Usage

just use your client to request. it servers on stdio

## Todo

- searchPackage
  - [ ] imported by how many packages
- getPackageInfo
  - [ ] get examples
- release
  - [x] use github actions to release for multiple platforms 

## Library Usage

The exported Go API of this module should currently be considered unstable, and subject to breaking changes. In the future, we may offer stability; please file an issue if there is a use case where this would be valuable.


## License

This project is licensed under the terms of the MIT open source license. Please refer to [MIT](./LICENSE) for the full terms.
