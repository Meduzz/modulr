# modulr
library to build dapr inspired proxies (which turns everything into a rest call).

It's now sort of divided into 3 areas.

1. api, where we store all definitions of how the library expect things to look.
2. lib, where we store the non-negotiable logic of the library.
3. adapter, where we store a bunch of default implementations of the adapters that can be used to extend or replace the functionallity of the library.

The example/proxy code will give you an idea of how it can work.

Currently the lib lets you:

* Register, reregister & lookup services, with support for addresses and types. Storage of this data can be customizable.
* Forward http requests to the services regististered, where how the loadbalancer bit work and the actual request/response dance work can be completely customized by the implementor.
* Listen to and send events and also deliver events to registered services. How events are sent can be customizeable, currently there's a default nats adapter available. How events are delivered back to the consumer can also be customized, currently there's a simple http adapter available.

## TODO

* Switch logging to gloegg
* Event api (specially EventSupport) turned out a bit too big, find a way to slice it
