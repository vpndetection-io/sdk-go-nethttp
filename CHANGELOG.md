# Changelog

What each release changed for you, newest first. Each line is a commit's summary, linked to its full description and diff. Releases before 2.1.2 are described by their release commits.

## 2.1.28 - 2026-11-11

### Fixes

- Require vpndetection v5.13.2: Unlimited, the free plan, never the Unlimited tier ([`483da45`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/483da4545eb85ac9ee7002f728c454b1d4f4ee5a))

## 2.1.27 - 2026-11-11

### Fixes

- Require vpndetection v5.13.1: the free plan's daily trial, never a trial key ([`13b92a8`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/13b92a83e46b3bbfc7412760845473894c1dda25))

## 2.1.26 - 2026-11-11

### Features

- Require vpndetection v5.13.0: trial_expired on a trial key's answers ([`4e00877`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/4e00877d367f94c31956e660ce1728ed9e4d850e))

## 2.1.25 - 2026-11-11

### Fixes

- Require vpndetection v5.12.2: the Unlimited tier and its terms ([`e2ae19a`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/e2ae19a6768116f717d6f3653383ce3c0dfeb3fb))

## 2.1.24 - 2026-11-11

### Fixes

- Require vpndetection v5.12.1: plainer words on lookups without a key ([`a3f6285`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/a3f6285621b409046031ca9db8e1d24656b126be))

## 2.1.23 - 2026-11-11

### Features

- Require vpndetection v5.12.0: a lookup without a key answers more ([`f82bd63`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/f82bd635f5b9204ab8e53fb75d791c0314211ab2))

## 2.1.22 - 2026-11-08

### Fixes

- Require sdk-go v5.11.1: the download methods keep v5.10's signatures ([`5244ede`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/5244ede13f3cdf010620bc65361c20e77ab4f9be))

## 2.1.21 - 2026-11-07

### Features

- Require sdk-go v5.11.0: downloads take the build a webhook names ([`ac406f8`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/ac406f8db652c2cf805a557fdf43b59d02d4efd4))

## 2.1.20 - 2026-10-29

### Fixes

- Require sdk-go v5.10.2: a lowercase t or z read, negative per-call retries refused ([`27024b6`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/27024b68e08094906a250cfbadc98e7d88a09102))

## 2.1.19 - 2026-10-26

### Fixes

- Require sdk-go v5.10.1: re-pinned to the 2026.10.25 spec ([`c7340cf`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/c7340cf70eb8101c98bfdbfe1b813998115d4c0b))

## 2.1.18 - 2026-10-22

### Fixes

- Require sdk-go v5.10.0: Retry-After and date-time zones, Oauth.Register ([`29c52f2`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/29c52f2425a1a330d08ee68d8a2d3a09099ff468))

## 2.1.17 - 2026-10-20

### Features

- Require sdk-go v5.9.0: lookup results carry Location and Asn ([`1ea1172`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/1ea1172264604f547351bbbf1eb889d10acf17fa))

## 2.1.16 - 2026-10-19

### Fixes

- Require sdk-go v5.8.0: num_ports and num_proxies in the pubproxy detail ([`fe19fb2`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/fe19fb2d40300f31eb27dd00b32e4ec00532572a))

## 2.1.15 - 2026-10-18

### Fixes

- Require sdk-go v5.7.0: is_pubproxy and the pubproxy detail ([`786849c`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/786849c3ff1f7eafea96e8d8d08ee56719c4aac0))

## 2.1.14 - 2026-10-17

### Features

- Require sdk-go v5.6.0: OauthMetadata carries registration_endpoint ([`cc93b20`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/cc93b204dbf1660817924017f96a5c5801b0f2d4))

## 2.1.13 - 2026-10-16

### Fixes

- Require sdk-go v5.5.3: unreadable answers retried, the key trimmed ([`0974297`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/097429762adc76b54a1e3a2d187052febc6bc4e4))

## 2.1.12 - 2026-10-10

### Fixes

- Require sdk-go v5.5.2: the spec re-pinned to 2026.10.09 ([`f332d4e`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/f332d4e71545c79c892ff31b35063123533e79d4))

## 2.1.11 - 2026-10-10

### Fixes

- Require sdk-go v5.5.1: the middleware starts with only an API key ([`29374b8`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/29374b8bfc50d55c677ed4270de674ee3df5e248))

## 2.1.10 - 2026-10-04

### Features

- Require sdk-go v5.5.0: the authorization code sign-in ([`4e16e24`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/4e16e249310c624c7c71484c50bef4977b5d2b0d))

## 2.1.9 - 2026-10-04

### Fixes

- Require sdk-go v5.4.4, re-pinned to spec 2026.10.03 ([`45c3b2f`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/45c3b2f72f418fc6c544389c716ff5e3b7f6a5cd))

## 2.1.8 - 2026-09-29

### Fixes

- Require sdk-go v5.4.3: 26 more reserved ranges are answered locally ([`bb61c6d`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/bb61c6d7b820e340ec07f51c7e85d60dbfbbf868))

## 2.1.7 - 2026-09-28

### Fixes

- Require sdk-go v5.4.2: IPv4-mapped visitors are looked up, not waved through ([`01ab295`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/01ab29544c3196beec90b0a49fd871f89230e227))

## 2.1.6 - 2026-09-28

### Fixes

- Require sdk-go v5.4.1: builds on 32-bit platforms again ([`f15da52`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/f15da52261367ce69579358b0255e85a488fc410))

## 2.1.5 - 2026-09-27

### Features

- Require sdk-go v5.4.0: OauthMetadata carries client_id_metadata_document_supported ([`6c6fb1c`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/6c6fb1c269e54b9e2810dd8006afff9fdd4c953b))

## 2.1.4 - 2026-09-27

### Fixes

- Require sdk-go v5.3.4: bounded Retry-After and poll, no doubled slash ([`63c82d3`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/63c82d39cdfddba94132ddcf88ffa6fcc18f6d94))

## 2.1.3 - 2026-09-26

### Fixes

- Require sdk-go v5.3.3: DownloadBytes no longer panics ([`389c119`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/389c119cf4cdc2101e6644ff075d541d0db95985))

## 2.1.2 - 2026-09-25

### Fixes

- Require sdk-go v5.3.2 ([`24d5825`](https://github.com/vpndetection-io/sdk-go-nethttp/commit/24d58255fd76ff9620214843dfbb1e149f016573))
