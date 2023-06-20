# cf-ip

**cf-ip** is a tool designed to scan IP addresses and identify Cloudflare reverse proxy IPs.

## Usage

To use **finder**, you need to specify the following required options:

```
Options:
  -asn int
        ASN (required)
  -n int
        Concurrency, default 10 (default 10)
  -s string
        Server name (required)
  -t_cf int
        CF proxy check timeout, default 3000(ms) (default 3000)
  -t_p int
        Port check timeout, default 500(ms) (default 500)
```

Example usage:

```bash
./finder -s example.com -asn 12345 -n 20 -t_cf 5000
```

This example command will scan the IP addresses associated with the server name "example.com" and the specified ASN 12345, using a concurrency level of 20. The timeout for CF proxy check is set to 5000 milliseconds.

Please note that the options `-s` and `-asn` are required, and you must provide values for both of them.

## TODO

- [ ] Support configuration file
- [ ] Enable log output
- [ ] Support custom IP addresses
- [ ] Implement network speed and latency testing
- [ ] Integrate with HAProxy

## Contribution

I’m a beginner in Golang and I’m still learning how to code better. I invite experienced developers to join me in improving cf-ip. I would appreciate any pull requests, feedback, or guidance from you.