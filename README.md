# kubernetes secrets manager

## <WIP>

### Available features

- connect to a kubernetes cluster choose a secret from a namespace and update its data
- save the secret to filesystem and enrypt it with SOPS

## What is it?

A TUI to create, update and encrypt kubernetes secrets.

## How to use?

Configure your kubernets cluster context in your terminal e.g. with [kubectx](https://github.com/ahmetb/kubectx).
Run the command to start the TUI, there you can choose the namespace which contains the secret and open it in a textfield.
Make your changes and hit ctrl+s to save and encrypt it with [SOPS](https://github.com/getsops/sops)
Upload is inteded to be done via some GitOps Tool (e.g. FluxCD etc)
