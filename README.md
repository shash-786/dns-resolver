# DNS Resolver

A lightweight DNS resolver implementation in Go that performs recursive DNS resolution starting from root DNS servers.

## Features

- Recursive DNS resolution
- Support for DNS message parsing and packing
- Uses root DNS servers for initial queries
- Handles multiple types of DNS records
- Concurrent request handling

## Prerequisites

- Go 1.18 or higher
- `golang.org/x/net` package

## Installation

```bash
git clone https://github.com/shash-786/dns-resolver
cd dns-resolver
go mod download
```

## Usage

Run the DNS server:
```bash
sudo go run main.go
```
Note: Root privileges are required as the server listens on port 53 (standard DNS port).

## How It Works

1. Server listens for incoming DNS queries on UDP port 53
2. Each query is handled in a separate goroutine
3. Resolution starts with root DNS servers
4. Follows referrals through the DNS hierarchy until finding an authoritative answer
5. Returns the resolved DNS response to the client

## Architecture

- `main.go`: Entry point, sets up UDP server
- `resolver.go`: Core DNS resolution logic including:
  - Query handling
  - Recursive resolution
  - DNS message processing
