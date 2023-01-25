# Replicated Security Persistent Testnet

The Replicated Security (RS) persistent testnet will be used to test Interchain Security features.

This testnet includes the following:
* A provider chain
* Consumer chains
* Relayers
* Block explorers
* Faucets

## Status

### Live Chains

* Provider: [`provider-1`](/replicated-security/provider-1/README.md)
* Consumer: [`timeout-2`](/replicated-security/timeout-2/README.md) (Waiting to timeout on `vsc_timeout_period`)
* Consumer: [`timeout-3`](/replicated-security/timeout-2/README.md) (Waiting to timeout on `ccv_timeout_period`)

### Stopped Chains

* Consumer: [`timeout-1`](/replicated-security/timeout-1/README.md) (Timed out on provider `init_timeout_period`)

## Upcoming Events

* 2023-01-26: [`consumer-1`](/replicated-security/consumer-1/README.md) chain launch
* 2023-01-30: `slasher-1` chain launch

See the [RS testnet schedule](SCHEDULE.md) for consumer chain launches and other planned events.
