# Contributing

## Reporting Issues

You can [open a new issue at GitHub](https://github.com/go-mclib/protocol/issues/new). Make it as descriptive as possible, along with a code sample or steps to reproduce the issue. The more information you provide, the better.

## Submitting Pull Requests

If you have improved the library, you can [open a new pull request at GitHub](https://github.com/go-mclib/protocol/pulls).

## Project Philosophy

This is intentionally a low-level library. A higher-level API is available in [go-mclib/client](https://github.com/go-mclib/client). The library follows a "live in the moment" policy - only the latest Minecraft Java Edition version is supported.

Key principles:

- Keep the API minimal; avoid legacy code and unnecessary abstractions
- Assume modern Go (1.25+) APIs
- Breaking changes are allowed (the higher-level packages will adapt)
- Prefer composable, generic patterns over embedded serialization logic

## Codebase

The codebase is written in Go, formatted with `go fmt`. It is divided into the following packages:

- [`auth`](./auth/): Microsoft OAuth2 authentication flow (login → Xbox Live → XSTS → Minecraft), session caching, and [Mojang certificate](https://minecraft.wiki/w/Mojang_API#Certificates) fetching for chat signing. [Learn how to obtain a client ID here](https://minecraft.wiki/w/Microsoft_authentication#Microsoft_OAuth2_flow);
- [`crypto`](./crypto/): SHA1 hash generation, CFB8 mode implementation, and AES encryption utilities used by the protocol;
- [`nbt`](./nbt/): Named Binary Tag format implementation with support for both file and network formats, struct marshaling, and a streaming visitor API;
- [`java_protocol`](./java_protocol/): Core Java Edition protocol implementation including:
  - [`java_protocol/net_structures`](./java_protocol/net_structures/): Protocol data types (`VarInt`, `VarLong`, `UUID`, `Position`, composite types like `PrefixedArray`, `XOrY`, etc.);
  - [`java_protocol/session_server`](./java_protocol/session_server/): Communication with [Mojang's session server](https://minecraft.wiki/w/Mojang_API#Verify_login_session_on_client) for authentication verification;
  - Packet serialization/deserialization with compression and encryption support;
  - TCP client/server connection handling with SRV record resolution;

For packet mappings, see [go-mclib/data](https://github.com/go-mclib/data).

## Code Style

- Comments should be minimal and only document non-obvious parts
- Comments (except docstrings or exported symbol names) should start with lowercase
- References to the [Minecraft Wiki](https://minecraft.wiki/w/Java_Edition_protocol) are appreciated where applicable

## Testing

Unit tests are located in separate `_test` packages. Tests validate functionality and packet mappings against actual packet dumps. Run all tests with:

```bash
go test ./...
```
