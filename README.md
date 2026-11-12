# go-mclib/protocol

Minimal Go bindings for Minecraft Java Edition 26.3 (protocol 777).

For packet mappings, see [go-mclib/data](https://github.com/go-mclib/data/tree/main/packets).

Note that this is a low-level library. It is used by the [go-mclib/client](https://github.com/go-mclib/client) repository to implement the client side of the protocol with a higher-level API. If you are looking to create a Minecraft (CLI) client, you should check that repository instead.

Also note that this library is not yet stable. The API is subject to change. The project is currently maintained by me ([@SKevo18](https://github.com/SKevo18)) and is not yet ready for production use. I will implement new features and bug fixes as my spare time allows, but I cannot guarantee that this project will not be left unmaintained.

## Dependency Chain

**[go-mclib/protocol](https://github.com/go-mclib/protocol)** <–––(requires)––– [go-mclib/data](https://github.com/go-mclib/data) <–––(requires)––– [go-mclib/client](https://github.com/go-mclib/client)

## Using the library in your project

The main branch supports only Minecraft 26.3. Install the current revision:

```bash
go get github.com/go-mclib/protocol@latest
```

Dependencies are resolved from `go.mod`; a local `go.work` file is optional.

### Text components

`wire.TextComponent` supports JSON and network NBT components, including
Minecraft 26.3 player-head and atlas-sprite objects. Use `TextComponent.Object`
with an `ObjectComponent`; its fields are flattened into the component on wire.
`ObjectComponent.Player` accepts a profile-name string or a `map[string]any` of
profile fields. Profile UUIDs use four `int32` values.

Object `Fallback` is a full text component; translation `Fallback` remains a
string. Plain-text rendering uses an object's explicit fallback or vanilla's
default description, preserving surrounding text. Profile names and textures
are decoration, not authenticated player identity.

## Quick Q&A

### Is there a simpler API?

A higher-level API is available in the [go-mclib/client](https://github.com/go-mclib/client) repository. It is a work in progress, but it is a good starting point for those who want to create a simple Minecraft client/bot in Go.

### Where can I learn more about the internals of the Minecraft protocol?

I can't recommend the [Minecraft Wiki](https://minecraft.wiki/w/Java_Edition_protocol) enough. The folks there did an amazing job at dissecting and documenting the protocol in as much detail as possible, which is a respect-worthy feat. I recommend that you read some of the pages there to get a better understanding of the protocol.

### How do I use this library as is (at the low-level)?

Check [go-mclib/client:example_bot.go](https://github.com/go-mclib/client/blob/938d84503cc368f59024a4911586b79ee943204a/example_client.go) for a low-level example of how to use this library. This code was the first ever working AFK bot (capable of joining a server and staying idle) that was made with this library. This also means that the code is messy, however, it works. It should give you a good idea of what the library is capable of and how to use the Minecraft protocol to create a bare-minimum bot.

### Is Bedrock Edition supported?

Not at the moment, but maybe in the future through a separate package (e. g. `bedrock_protocol`).

### Why is Minecraft version `X` not supported?

Since I am the only one maintaining this project at the moment, adding concurrent support for older Minecraft versions is practically impossible, also given the limited spare time I have. Therefore, the library has a "live in the moment" policy, where only the latest version of the Minecraft Java Edition protocol is supported.

Older protocol versions are not supported by the current branch.

## Contributing

Interested in learning more about the internals, or contributing to the project? Check out the [CONTRIBUTING.md](./CONTRIBUTING.md) file.

## License

This project is licensed under the [MIT License](./LICENSE).md.

## Donate

If you want to support my work (thanks!), you can [sponsor me on GitHub](https://github.com/sponsors/SKevo18).
