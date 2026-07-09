# SecSnapGo — Live Forensic Snapshot Daemon

A Golang port of Python-based background daemon that monitors system activity in real time and automatically captures forensic snapshots when suspicious behavior is detected. Built for SOC and DFIR workflows.

## Features
- Real-time system monitoring across CPU, RAM, network, and disk
- Threshold-based anomaly detection with configurable triggers
- Configurable whitelist for trusted IPs and processes to reduce false positives
- Email alerting when a forensic snapshot is triggered
- Dual format report generation (TXT + JSON)
- Cooldown mechanism to prevent snapshot flooding
- Automatic snapshots directory creation on first run

## What It Monitors
- CPU usage per core and frequency
- RAM consumption and top memory processes
- Active network connections and outbound connections to known malicious ports
- Disk I/O rates and recently modified files in /tmp

## Trigger Conditions
- CPU usage exceeds threshold (default 85%)
- RAM usage exceeds threshold (default 80%)
- Outbound connection to suspicious port (4444, 1337, 31337, 9001, 6667)
- Disk write rate exceeds threshold (default 50 MB/s)

## Whitelist Configuration
To reduce false positives, you can configure trusted IPs and processes that should be ignored during monitoring:

- Whitelisted IPs — Outbound connections to these IPs will not trigger snapshots
- Whitelisted Processes — System or trusted processes can be excluded from triggering alerts
- Whitelisted Ports — Safe ports that should not be treated as suspicious

Edit config.py to define whitelists with examples like localhost IPs, system processes, and common web ports.

## Output
Each triggered snapshot generates two timestamped files in snapshots/ (auto-created if missing):
- snapshot_YYYYMMDD_HHMMSS.txt — human-readable forensic report
- snapshot_YYYYMMDD_HHMMSS.json — structured data for downstream tooling

## Project Structure

* #### config.go #all whitelists and settings/thresholds
* #### Utils.go collectors,snaphots, etc.
* #### main.go main program

## Usage
```bash
go run main .go
```
#### or

```bash
go build .
./main
```

## Configuration
Edit config.go to adjust thresholds, whitelists, and email settings including CPU threshold, RAM threshold, disk write threshold, suspicious ports list, whitelisted IPs, whitelisted processes, whitelisted ports, email alerts toggle, daemon interval, and cooldown seconds.

## Setup

install dependencies with
```bash
go mid tidy 
```

## Skills Demonstrated
- Golang daemon architecture
- Real-time system monitoring with golang psutil
- Forensic data collection across CPU, RAM, network, and disk
- Threshold-based anomaly detection
- Whitelist filtering for false positive reduction
- Email alerting integration for SOC workflows
- Dual format report generation (TXT + JSON)
- DFIR and SOC-relevant incident response workflows

## Author
* mwcsur@gmail.com / midwestcoder2020
* GitHub: https://github.com/midwestcoder2020
