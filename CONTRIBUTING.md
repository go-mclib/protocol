# Contributing

## Reporting Issues

You can [open a new issue at GitHub](https://github.com/go-mclib/protocol/issues/new). Make it as descriptive as possible, along with a code sample or steps to reproduce the issue. The more information you provide, the better.

## Submitting Pull Requests

If you have improved the library, you can [open a new pull request at GitHub](https://github.com/go-mclib/protocol/pulls).

### Codebase

The codebase is written in Go, formatted with `go fmt`. It is divided into the following packages:

- [`auth`](./auth/): Handles authentication through Microsoft's OAuth2 flow. [Learn how to obtain a client ID here](https://minecraft.wiki/w/Microsoft_authentication#Microsoft_OAuth2_flow);
- [`crypto`](./crypto/): Contains utilities for cryptography (SHA1 hash generation, CFB8 implementation, network encryption, etc.);
- [`java_protocol`](./java_protocol/): Contains core utilities for the Java Edition protocol, like packet serialization and deserialization;
- [`java_protocol/packets`](./java_protocol/packets/): Concrete packet structs for the given branch (currently: `772` for 1.21.8);
- [`net_structures`](./net_structures/): Implementation and parsing for network structures used in Minecraft protocol (`VarInt`, `VarLong`, `UUID`, etc.);
- [`session_server`](./session_server/): Handles communication with [Mojang's session server](https://minecraft.wiki/w/Mojang_API#Verify_login_session_on_client) (`sessionserver.mojang.com`);

Try to make the codebase relatively readable, modular, easy to understand/maintain and test. The code should be well-documented, with references to the Minecraft wiki where applicable.

### Testing

The codebase is tested with [Go's built-in testing framework](https://go.dev/doc/testing). The tests are located in a separate package (ending with `_test`). You can run all of them with `go test ./...`.
