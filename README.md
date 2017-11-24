# airworthy

Download and verify binaries based on GPG signatures

* embeds publicly trusted ca-certificates for HTTPS
* embeds GPG public keys for Jetstack and HashiCorp releases

## Usage

### Build

```
go generate ./pkg/... && go build
```

### Download and verify

```
./airworthy -v download https://github.com/jetstack/vault-helper/releases/download/0.9.2/vault-helper_0.9.2_linux_amd64                                                                                                                                                                                               1 ↵
DEBU[0000]                                               flags=&{  https://github.com/jetstack/vault-helper/releases/download/0.9.2/vault-helper_0.9.2_linux_amd64  <nil>}
DEBU[0000] set signature to URL                          signature="https://github.com/jetstack/vault-helper/releases/download/0.9.2/vault-helper_0.9.2_linux_amd64.asc"
DEBU[0000] set destination                               destination="vault-helper_0.9.2_linux_amd64"
DEBU[0000] keyring contains: id=9A1C42C8F5AA3CE6 (Jetstack Releases <tech+releases@jetstack.io>)
DEBU[0000] keyring contains: id=51852D87348FFC4C (HashiCorp Security <security@hashicorp.com>)
INFO[0000] downloading https://github.com/jetstack/vault-helper/releases/download/0.9.2/vault-helper_0.9.2_linux_amd64...
INFO[0001]   200 OK
DEBU[0001]   transferred 243263 / 6998496 bytes (3.48%)
DEBU[0002]   transferred 2627063 / 6998496 bytes (37.54%)
INFO[0002] download saved to ./vault-helper_0.9.2_linux_amd64
INFO[0002] successfully signed by id=9A1C42C8F5AA3CE6 (Jetstack Releases <tech+releases@jetstack.io>)
```
