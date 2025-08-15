# gomc-protocol

Minimal Go bindings for the Minecraft server protocol.

For packet mappings, see [go-mclib/data](https://github.com/go-mclib/data/tree/main/go/772/java_packets)

Note that this is a low-level library. It is used by the [go-mclib/client](https://github.com/go-mclib/client) repository to implement the client side of the protocol with a higher-level API. If you are looking to create a Minecraft (CLI) client, you should check that repository instead.

Also note that this library is not yet stable. The API is subject to change. The project is currently maintained by me ([@SKevo18](https://github.com/SKevo18)) and is not yet ready for production use. I will implement new features and bug fixes as my spare time allows, but I cannot guarantee that this project will not be left unmaintained.

## Using the library in your project

This project is still in its early stages, so there are no tagged releases yet. You can use the `@<protocol-version>` branch to get the latest changes. For example:

```bash
go get github.com/go-mclib/protocol@772
```

Replace `772` with the protocol version (branch) you want to use.

## Quick Q&A

### Is there a simpler API?

A higher-level API is available in the [go-mclib/client](https://github.com/go-mclib/client) repository. It is a work in progress, but it is a good starting point for those who want to create a simple Minecraft client/bot in Go.

### Where can I learn more about the internals of the Minecraft protocol?

I can't recommend the [Minecraft Wiki](https://minecraft.wiki/w/Java_Edition_protocol) enough. The folks there did an amazing job at dissecting and documenting the protocol in as much detail as possible, which is a respect-worthy feat. I recommend that you read some of the pages there to get a better understanding of the protocol.

### How do I use this library as is (at the low-level)?

Check [go-mclib/client:example_bot.go](https://github.com/go-mclib/client/blob/938d84503cc368f59024a4911586b79ee943204a/example_client.go) for a low-level example of how to use this library. This code is the first ever working AFK bot (capable of joining a server and staying idle) that was made with this library. This also means that the code is messy, however, it works. It should give you a good idea of what the library is capable of and how to use the Minecraft protocol to create a bare-minimum bot.

### Is Bedrock Edition supported?

Not at the moment, but maybe in the future through a separate package (e. g. `bedrock_protocol`).

### Why is Minecraft version `X` not supported?

Since I am the only one maintaining this project at the moment, adding concurrent support for older Minecraft versions is practically impossible, also given my limited spare time constraints. Therefore, the library has a "live in the moment" policy, where a branch resembles a specific protocol version. For example, the `772` branch is the latest version of the library that supports Minecraft: Java Edition 1.21.8 and was only tested on that version.

It is possible that the library might work for older/newer versions (assuming that there were no breaking changes in terms of the networking or packet structures), but I cannot guarantee that. Feel free to experiment and open a PR to help making the library more versatile.

## Contributing

Interested in learning more about the internals, or contributing to the project? Check out the [CONTRIBUTING.md](./CONTRIBUTING.md) file.

## License

This project is licensed under the [MIT License](./LICENSE).md.

## Donate

If you want to support my work (thanks!), you can [sponsor me on GitHub](https://github.com/sponsors/SKevo18).
