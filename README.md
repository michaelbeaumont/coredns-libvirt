# coredns-libvirt

This plugin allows CoreDNS to resolve names matching libvirt domains. It is
similar to the `nss` plugin `libvirt_guest`.

## Usage

Currently this plugin can only be used in the following mode.

### Guest name

The functionality of `libvirt guest` is analogous to the `libvirt_guest` `nss`
plugin, where we look for a match on the name of the libvirt domain, not
necessarily a hostname.

### Filtering by network

If only some of the IPs assigned to the guests are reachable, you can filter
them with the `keep` directive.

## Example

```
vm.:1053 {
  rewrite name suffix .vm .
  libvirt guest {
    keep 10.101.0.0/24
  }
}
```
