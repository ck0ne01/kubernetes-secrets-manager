# kubernetes secrets manager

Ksm provides a terminal UI to handle kubernetes secrets. The aim of this project is to make it easier to change existing encrypted secrets specially with many data entries.
It let you choose an existing secret in a kubernetes cluster, edit the data and write it to an encrypted file.

## Installation

Download a pre-built binary for your architecture from the realeases page.

## Usage

### Context

Configure your kubernets cluster context in your terminal kubectl config use-context <CONTEXT>.
The KUBECONFIG environment variable is currently not supportet, to use this TUI you can write a new config file.

```
kubectl view --flatten > PATH-TO-CONFIG-FILE
```

Then use the `--kubeconfig=PATH-TO-CONFIG-FILE` arg to specify the newly created config.

### Encryption

Encryption is done with [SOPS](https://github.com/getsops/sops) and thus is needed to be installed and configured.
Follow the instructions in the repo to set it up.

### Create a secret

![create secret](docs/create-secret.gif)

### Update a secret

![update secret](docs/update-secret.gif)
