<!--
order: 1
-->

# Tendermint KMS

Set up a Key Management System for Evmos {synopsis}

[Tendermint KMS](https://github.com/iqlusioninc/tmkms) is a Key Management Service (KMS) that allows separating key management from Tendermint nodes. In addition it provides other advantages such as:

- Improved security and risk management policies
- Unified API and support for various HSM (hardware security modules)
- Double signing protection (software or hardware based)

It is recommended that the KMS service runs in a separate physical hosts.

## Install Tendermint KMS onto the node

You will need the following prerequisites:

- ✅ **Rust** (stable; **1.56+**): https://rustup.rs/
- ✅ **C compiler**: e.g. gcc, clang
- ✅ **pkg-config**
- ✅ **libusb** (1.0+). Install instructions for common platforms
    - ✅ Debian/Ubuntu

      ```bash
      apt install libusb-1.0-0-dev
      ```

    - ✅ RedHat/CentOS
  
      ```bash
      yum install libusb1-devel
      ```

    - ✅ macOS (Homebrew)
  
      ```
      brew install libusb
      ```

::: tip
For `x86_64` architecture only:

Configure `RUSTFLAGS` environment variable:

```bash
export RUSTFLAGS=-Ctarget-feature=+aes,+ssse3
```

:::

We are ready to install KMS. There are 2 ways to do this: compile from source or install with Rusts cargo-install. We’ll use the first option.

### Compile from source code

The following example adds `--features=ledger` to enable Ledger  support.
`tmkms` can be compiled directly from the git repository source code, using the following commands:

```bash
gh repo clone iqlusioninc/tmkms && cd tmkms
[...]
cargo build --release --features=ledger
```

Alternatively, substitute `--features=yubihsm` to enable [YubiHSM](https://www.yubico.com/products/hardware-security-module/) support.

If successful, it will produce the `tmkms` executable located at: `./target/release/tmkms`.

## Configuration

A KMS can be configured using the following HSMs

### YubiHSM
  
Detailed information on how to setup a KMS with [YubiHSM 2](https://www.yubico.com/products/hardware-security-module/) can be found [here](https://github.com/iqlusioninc/tmkms/blob/master/README.yubihsm.md).

### Ledger Tendermint app

Detailed information on how to setup a KMS with Ledger Tendermint App can be found [here](kms_ledger.md).
