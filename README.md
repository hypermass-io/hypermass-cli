# Hypermass CLI
The official command-line interface for **Hypermass**, the high-performance data distribution platform.

Hypermass is designed to distribute large files at low latency. Use this CLI tool to subscribe and publish your data.

## Quick Start
You can download the latest pre-compiled binaries for your operating system from the [Releases Page](https://github.com/hypermass-io/hypermass-cli/releases).

These can be run directly in the terminal, or installed - Installation instructions here [Installation Guide](https://docs.hypermass.io/docs/cli/download-and-install).

## Subscribing to Data
Copy a stream's ID from its page on [hypermass.io](https://hypermass.io), then;
```bash
hypermass subscribe <stream id>
hypermass sync
```
Files arrive in `<home folder>/hypermass/subscriptions/<stream id>` as they are published. Subscribing works straight away, within a
free daily allowance per address.

## Signing in
An access key from your account gives you a larger allowance and lets you publish. Create one at
https://hypermass.io/access-keys, then;
```bash
hypermass login
```

## Publishing Data
```bash
hypermass publish <stream id>
```
Files you place in `<home folder>/hypermass/publications/<stream id>` are published to the stream, then deleted.

## Configuration
The hypermass-config.yaml configuration file says what to subscribe and publish to, and where the files go. You can
print its location with;
```bash
hypermass info
```
Edit it however you need, then apply the changes to a running sync with `hypermass reload`. Full configuration guide
here [Configuration Guide](https://docs.hypermass.io/docs/cli/configuration).

You may want to back up the hypermass-config.yaml file used as part of deployments (e.g. in a git repo) - it's plain
text and contains no security details. The access key lives separately, in auth.yaml, which we advise against backing
up; a new key is easy to create at https://hypermass.io/access-keys.

## Key Features
* **File-Based Configuration:** Human-readable YAML setup. Easy to back up, version control, etc. No complex database or registry entries required.
* **Atomic File Delivery:** The "Write-and-Move" strategy ensures that if a file appears in your target folder, it is 100% complete and verified. 
* **Flexible Receiver Strategies:** Choose between `file-per-payload` for simplicity or `folder-with-metadata` for rich data handling.
* **Production-Grade Security:** Native SSL support and secure token-based authentication out of the box.

## Documentation
Full documentation can be found at [docs.hypermass.io](https://docs.hypermass.io).

## License
Distributed under the **Apache 2.0 License**. See `LICENSE` for more information.

---
[hypermass.io](https://hypermass.io) | [@hypermass_io](https://twitter.com/hypermass_io)
