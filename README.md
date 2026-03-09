# Hypermass CLI
The official command-line interface for **Hypermass**, the high-performance data distribution platform.

Hypermass is designed to distribute large files at low latency. Use this CLI tool to subscribe and publish your data.

## Quick Start
You can download the latest pre-compiled binaries for your operating system from the [Releases Page](https://github.com/hypermass-io/hypermass-cli/releases).

These can be run directly in the terminal, or installed - Installation instructions here [Installation Guide](https://docs.hypermass.io/docs/cli/download-and-install).

## Initialising the configuration
You will need to initialise the configuration directory on the first run. 
This will prompt you to enter credentials from your account (you can create one here if you have signed up: https://hypermass.io/access-keys).

Note: if you have a stream in mind to test with, then grab the "key" from the subscribe page of that stream and have it 
ready. You can drop it into "init" and get started faster.
```bash
hypermass init
```
Input the appropriate data as prompted.

You can print out the configuration location with this command;
```bash
hypermass info
```

The hypermass-config.yaml configuration file tells the hypermass cli command what to subscribe to and where to put the
result. You can now edit the "hypermass-config.yaml" however you need. Full configuration guide here [Configuration Guide](https://docs.hypermass.io/docs/cli/configuration).
You may want to back up the hypermass-config.yaml files used as part of deployments (e.g. in a git repo) - it's plain text and contains no security details.

We advise leaving the auth.yml alone unless you want to change keys. We advise against backing the key value up for 
security reasons, but it's easy enough to generate a new key here: https://hypermass.io/access-keys.

### Subscribing to Data
If you added a stream key in the init command you're all set (or subsequently configured it, per the [Configuration Guide](https://docs.hypermass.io/docs/cli/configuration)), just run;
```bash
./hypermass sync
```

Streams that you are subscribed to will appear in;
    <<HOME>>/hypermass/data/subscribe/hypermass-status
    <<HOME>>/hypermass/data/subscribe/arbitrary-name

## Key Features
* **File-Based Configuration:** Human-readable YAML setup. Easy to back up, version control, etc. No complex database or registry entries required.
* **Atomic File Delivery:** The "Write-and-Move" strategy ensures that if a file appears in your target folder, it is 100% complete and verified. 
* **Flexible Receiver Strategies:** Choose between `file-per-payload` for simplicity or `folders-with-metadata` for rich data handling.
* **Production-Grade Security:** Native SSL support and secure token-based authentication out of the box.

## Documentation
Full documentation can be found at [docs.hypermass.io](https://docs.hypermass.io).

## License
Distributed under the **Apache 2.0 License**. See `LICENSE` for more information.

---
[hypermass.io](https://hypermass.io) | [@hypermass_io](https://twitter.com/hypermass_io)
