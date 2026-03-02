# Running as a Background Service

If you are using the Hypermass CLI on a server or a persistent workstation, you’ll likely want the sync process to run 
automatically in the background and restart if the system reboots.

## Recap
The "hypermass sync" command will synchronise streams to local folders. The specific streams and location of the folders
is controlled in the yaml configuration. It works for both uploading data to streams (publishing) and downloading data 
from streams (subscribing).

For server applications, the "sync" command is how we expect most users to automatically integrate to Hypermass - it 
solves a lot of problems (e.g. retries, backoff, meta data communications, authentication, validation, hashing, etc) 
that would otherwise be a pain to implement.

See the main docs about [how to implement a CLI client](https://docs.hypermass.io/docs/cli/building-a-client)

## Please note
We've tested the Linux approach, we'd appreciate feedback on the Windows and macOs instructions. 

Specifically with the Windows and macOs instructions, please take this cautiously and verify. 

If anything is wrong either in the instructions or missed edge cases, please raise an Issue.


## Linux (SystemD)
This covers the common modern options: RHEL & Derivatives, Ubuntu, SUSE & Derivatives, Amazon Linux 2023, Arch, etc

```bash
sudo nano /etc/systemd/system/hypermass.service
```

Paste the following configuration (adjust /usr/local/bin/hypermass if your binary is elsewhere):
Ini, TOML

```systemd.syntax
[Unit]
Description=Hypermass Sync Daemon
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
Restart=always
RestartSec=5
User=your-username
ExecStart=/usr/local/bin/hypermass sync
WorkingDirectory=/home/your-username

[Install]
WantedBy=multi-user.target
```

Enable and start the service:
```BASH
    sudo systemctl daemon-reload
    sudo systemctl enable hypermass
    sudo systemctl start hypermass
```

## macOS (launchd)
*** Untested! Please confirm/correct ***

On macOS, we use launchd to manage background agents.

Create a "plist" file in your user library:
```Bash

nano ~/Library/LaunchAgents/io.hypermass.sync.plist

Paste the following:
XML

<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>io.hypermass.sync</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/hypermass</string>
        <string>sync</string>
    </array>
    <key>KeepAlive</key>
    <true/>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>
```

Load the agent:
```Bash

    launchctl load ~/Library/LaunchAgents/io.hypermass.sync.plist
```


## Windows (PowerShell/Scheduled Task)

*** Untested! Please confirm/correct ***

For a "headless" background experience on Windows without a complex service wrapper:

Open PowerShell as Administrator.

Run the following to create a task that starts hypermass at every user login:

```PowerShell
$action = New-ScheduledTaskAction -Execute "C:\Path\To\hypermass.exe" -Argument "sync"
$trigger = New-ScheduledTaskTrigger -AtLogOn
Register-ScheduledTask -Action $action -Trigger $trigger -TaskName "HypermassSync" -Description "Keep local folders synced with Hypermass nodes."
```

## Monitoring Logs
When running in the background, you can check your activity here:

Linux: journalctl -u hypermass -f
macOS: Check your local log file if redirected, or use log show --predicate 'process == "hypermass"'
Windows: Check the "History" tab in Task Scheduler.
