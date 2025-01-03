# btcleaner

Keep a minimum free space on your Transmission daemon while ensure a minimumm torrents per tracker to not being flag inactive.

## Configuration

Create a config file like this in ~/config/btcleaner.yaml

```yaml
transmission_url: "http://user:password@127.0.0.1:9091/transmission/rpc"

path: "/path/to/download/dir"
free_giga: 100
tracker_keep: 2

exclude:
  - torrent name
  - another torrent name
```

## Usage

Flags for `free_giga` and `tracker_keep` are used to temporary force value without edit config

```
Usage of btcleaner:
  -c, --config string      Config file path
      --cron               Same as --do flag but no ouput when enough free space
      --do                 Commit torrent deletion
  -f, --free_giga int      Target GiB free (default 250)
  -h, --help               This help message
  -k, --tracker_keep int   Torrent to keep per tracker (default 2)
  -v, --version            This help message
```

Usage example

```
$ btcleaner --do
141.78 GiB left, target is 150.00 GiB free. We need to recover  8.22 GiB .
---
Excluded because only 2 torrents left on  <masked> :  VideoReDo TVSuite H.264 v5.1.1.719b
Excluded because only 1 torrents left on  <masked> :  Ron's.Gone.Wrong.2021.MULTi.VFF.1080p.mHD.x264.AC3-XSHD.mkv
Excluded because only 2 torrents left on  <masked> :  In Flames AAC 320
Recovered 7.37 GiB / 8.22 GiB : ( 7.37 GiB ) [ <masked> ] Starship Troopers (1997) MULTi VFF 2160p 10bit 4KLight DV HDR BluRay DDP 5.1 Atmos x265-QTZ.mkv
Recovered 11.60 GiB / 8.22 GiB : ( 4.24 GiB ) [ <masked> ] Princesse Mononoké.mkv
2 torrent deleted.  153.38 GiB Free.
```
