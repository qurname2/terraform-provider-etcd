# Terraform Provider Etcd

This project has been based on https://github.com/hashicorp/terraform-provider-hashicups

Its allow you to manage some etcd elements via terraform.

## How to use it?

Please check [/docs](https://github.com/qurname2/terraform-etcd-provider/tree/main/docs) directory.

## Requirements
* [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.3.x
* [Go](https://go.dev/doc/install) >= 1.18

## Building The Provider
* Clone the repository
* Enter the repository directory
* Build the provider using the Go install command:
```bash
make build
```
### Installing the Provider
```bash
make install
```
## TODO list
* write tests for all resources
