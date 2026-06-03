## MarsChain Build

### Pull code
```bash
git clone https://github.com/marschain-protocol/marschain
```
### Build locally on macOS

Need to install Go 1.22. Note: higher versions are not supported.

```bash
brew install go@1.22
go version
```

Compile MarsChain Linux x64 version:

```bash
make marschain-linux-x64
```

### Server requirements

- 4 CPU cores
- 16 GB memory
- 1 TB SSD
- Port `18555` allowed through the firewall

### Copy files to the server

Copy the following files to a folder on the server:

- `marschain-linux-x64`
- `genesis.json`
- `config.toml`

### Initialize node

```bash
./marschain-linux-x64 init genesis.json --datadir ./node
```

### Create marschain service

Create file `/etc/systemd/system/marschain.service` with following content. Note:WorkingDirectory and ExecStart need to be replaced with real path where executable file located

```ini
[Unit]
Description=Start MarsChain Service
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/marschain
ExecStart=/opt/marschain/marschain-linux-x64 --config /opt/marschain/config.toml --datadir /opt/marschain/node --syncmode snap --cache 8000
RemainAfterExit=true
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
```

Then run:

```bash
sudo systemctl daemon-reload
sudo systemctl enable marschain.service
```

### Start marschain service

```bash
sudo systemctl start marschain.service
```

### Monitor log file

```bash
tail -f /opt/marschain/node/info.log
```