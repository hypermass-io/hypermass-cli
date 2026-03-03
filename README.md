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
./hypermass init
```
Input the appropriate data as prompted.

You can print out the configuration location with this command;
```bash
./hypermass info
```

You can now edit the "hypermass-config.yaml" however you need. Full configuration guide here [Installation Guide](https://docs.hypermass.io/docs/cli/configuration).
You may want to back up the hypermass-config.yaml files used as part of deployments (e.g. in a git repo) - it's plain text and contains no security details. 

We advise leaving the auth.yml alone unless you want to change keys. We advise against backing the key value up for 
security reasons, but it's easy enough to generate a new key here: https://hypermass.io/access-keys.

### Subscribing to Data
The hypermass-config.yaml configuration file tells the hypermass cli command what to subscribe to and where to put the 
result.

Here's a simple example that both publishes and subscribes to data;
```yaml
target-directory: /home/hypermass-user/data # default location of the subscriptions
subscription-targets:
  - key: "_WZBZ1QU" # the key of the stream
    target-directory: /home/hypermass-user/data/subscribe/hypermass-status # override the default target directory for this stream
    start-point: "latest" # where to start streaming when first connecting, default is latest. Allowed values: latest, earliest
    writer-type: "file-per-payload" # select "file-per-payload" for one file per payload, "folder-with-metadata" for a folder with metadata, default is "file-per-payload"
  - key: "_Cwj5GUPMF" # the key of the stream
    target-directory: /home/hypermass-user/data/subscribe/arbitrary-name # override the default target directory for this stream
    start-point: "latest" # where to start streaming when first connecting, default is latest. Allowed values: latest, earliest
    writer-type: "folders-with-metadata" # select "file-per-payload" for one file per payload, "folders-with-metadata" for a folder with metadata, default is "file-per-payload"

publication-sources:
  - key: "_E8P712QZ" # the key of the stream
    target-directory: /home/hypermass-user/data/publish/my-awesome-data-stream # override the default target directory for this stream
    disposer-type: "delete-on-success" # select "delete-on-success" or "move-on-success" as needed
```

If you added a stream key in the init command you're all set, just run;
```bash
./hypermass sync
```

Streams that you are subscribed to will appear in;
    /home/hypermass-user/data/subscribe/hypermass-status
    /home/hypermass-user/data/subscribe/arbitrary-name

And any files that you drop into:
    /home/hypermass-user/data/publish/my-awesome-data-stream
Will be read, uploaded to hypermass and deleted from the folder. Once in Hypermass the file will be verified and if it passes, published for other users to receive.

# Deploying on a server
On a sever you likely want this process running as a persistent service. See here for details: [Install hypermass sync as a service](install-sync-as-a-service.md)

## Key Features
* **File-Based Configuration:** Human-readable YAML setup. Easy to back up, version control, etc. No complex database or registry entries required.
* **Atomic File Delivery:** The "Write-and-Move" strategy ensures that if a file appears in your target folder, it is 100% complete and verified. 
* **Flexible Receiver Strategies:** Choose between `file-per-payload` for simplicity or `folders-with-metadata` for rich data handling.
* **Production-Grade Security:** Native SSL support and secure token-based authentication out of the box.

## Documentation
Full documentation can be found at [docs.hypermass.io](https://docs.hypermass.io).

## License
Distributed under the **MIT License**. See `LICENSE` for more information.

---
[hypermass.io](https://hypermass.io) | [@hypermass_io](https://twitter.com/hypermass_io)
