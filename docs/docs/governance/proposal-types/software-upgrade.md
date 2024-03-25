---
order: 2
parent:
  order: 2
---

# Software Upgrade

Software upgrade proposals are submitted to signal that a Cosmos Hub release with new features, bugfixes and various other improvements is available and ready for production deployment.

Software upgrade proposals should be submitted by the development teams tasked with stewarding the Cosmos Hub development.

## Procedure

Use `draft-proposal` command to create a draft proposal and populate it with required information.

```sh
✗ gaiad tx gov draft-proposal
Use the arrow keys to navigate: ↓ ↑ → ←
? Select proposal type:
    text
    community-pool-spend
  ▸ software-upgrade # choose this
    cancel-software-upgrade
    other

# populate all steps (displaying all for demonstration purposes)
Enter proposal title: Upgrade v15
Enter proposal authors: Stewards
Enter proposal summary: Upgrade to v15
Enter proposal details: <v15 upgrade changelog details>
Enter proposal proposal forum url: /
Enter proposal vote option context: Vote YES to support running this binary on the Cosmos Hub mainnet.
Enter proposal deposit: 100001uatom
Enter msg authority: cosmos10d07y265gmmuvt4z0w9aw880jnsr700j6zn9kn
```

In your `draft_proposal.json` populate the `height` with your desired upgrade height and populate the `info` field with additional information (must be a valid JSON string):
```json
{
  "binaries": {
    "darwin/amd64": "https://github.com/cosmos/gaia/releases/download/v15.0.0/gaiad-v15.0.0-darwin-amd64?checksum=sha256:7157f03fbad4f53a4c73cde4e75454f4a40a9b09619d3295232341fec99ad138",
    "darwin/arm64": "https://github.com/cosmos/gaia/releases/download/v15.0.0/gaiad-v15.0.0-darwin-arm64?checksum=sha256:09e2420151dd22920304dafea47af4aa5ff4ab0ddbe056bb91797e33ff6df274",
    "linux/amd64": "https://github.com/cosmos/gaia/releases/download/v15.0.0/gaiad-v15.0.0-linux-amd64?checksum=sha256:236b5b83a7674e0e63ba286739c4670d15d7d6b3dcd810031ff83bdec2c0c2af",
    "linux/arm64": "https://github.com/cosmos/gaia/releases/download/v15.0.0/gaiad-v15.0.0-linux-arm64?checksum=sha256:b055fb7011e99d16a3ccae06443b0dcfd745b36480af6b3e569e88c94f3134d3",
    "windows/armd64": "https://github.com/cosmos/gaia/releases/download/v15.0.0/gaiad-v15.0.0-windows-amd64.exe?checksum=sha256:f0224ba914cad46dc27d6a9facd8179aec8a70727f0b1e509f0c6171c97ccf76",
    "windows/arm64": "https://github.com/cosmos/gaia/releases/download/v15.0.0/gaiad-v15.0.0-windows-arm64.exe?checksum=sha256:cbbce5933d501b4d54dcced9b097c052bffdef3fa8e1dfd75f29b34c3ee7de86"
  }
}
```

Upload your `draft_metadata.json` to a distribution platform of your choice. `draft_proposal.json` is used to submit a governance proposal using `submit-proposal`.

```sh
gaiad tx gov submit-proposal <path_to_proposal.json>
   --from <submitter address> \
   --chain-id cosmoshub-4 \
   --gas <max gas allocated> \
   --fees <fees allocated> \
   --node <node address> \
```

## Using x/upgrading

Software upgrade proposals can be submitted using the [x/upgrade module](https://docs.cosmos.network/v0.47/build/modules/upgrade#transactions). The end effect will be the same since the `x/gov` module routes the message to `x/upgrade` module.

## Additional information

Additional instructions with debugging information is available on the [submitting](../submitting.md) page.

