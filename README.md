# gomc-protocol

Minimal Go bindings for the Minecraft server protocol.

Currently, it maps a handful of serverbound (C2S) packets for Minecraft: Java Edition 1.21.8 (protocol version 772). See the [`java_protocol/packets`](java_protocol/packets) package.

Note that this is a low-level library. It is used by the [go-mclib/client](https://github.com/go-mclib/client) repository to implement the client side of the protocol with a higher-level API. If you are looking to create a Minecraft (CLI) client, you should check that repository instead.

Also note that this library is not yet stable. The API is subject to change. The project is currently maintained by me ([@SKevo18](https://github.com/SKevo18)) and is not yet ready for production use. I will implement new features and bug fixes as my spare time allows, but I cannot guarantee that this project will not be left unmaintained.

## Contributing

Interested in learning more about the internals, or contributing to the project? Check out the [CONTRIBUTING.md](./CONTRIBUTING.md) file.

## License

This project is licensed under the [MIT License](./LICENSE).md.

## Donate

If you want to support my work (thanks!), you can [sponsor me on GitHub](https://github.com/sponsors/SKevo18).
