# Mechain Cosmos-SDK

This repo is forked from [cosmos-sdk](https://github.com/cosmos/cosmos-sdk).

The Mechain Block Chain leverages cosmos-sdk to fast build a dApp running with tendermint. Due to the many
requirements of Mechain blockchain that cannot be fully satisfied by cosmos-sdk at present, we have decided to fork
the cosmos-sdk repo and add modules and features based on it.

## Disclaimer
**The software and related documentation are under active development, all subject to potential future change without
notification and not ready for production use. The code and security audit have not been fully completed and not ready
for any bug bounty. We advise you to be careful and experiment on the network at your own risk. Stay safe out there.**

## Key Features

1. **auth**. The address format of the Mechain blockchain is fully compatible with BSC (and Ethereum). It accepts EIP712 transaction signing and verification. These enable the existing wallet infrastructure to interact with Mechain at the beginning naturally.
2. **crosschain**. Cross-chain communication is the key foundation to allow the community to take advantage of the Mechain and BSC (and Ethereum compatible) dual chain structure..
3. **gashub**. As an application specific chain, Mechain defines the gas fee of each transaction type instead of calculating gas according to the CPU and storage consumption.
4. **gov**. There are many system parameters to control the behavior of the Mechain and its smart contract on BSC (and Ethereum compatible), e.g. gas price, cross-chain transfer fees. All these parameters will be determined by Mechain Validator Set together through a proposal-vote process based on their staking. Such the process will be carried on cosmos sdk.
5. **oracle**. The bottom layer of cross-chain mechanism, which focuses on primitive communication package handling and verification.
6. **upgrade**. Seamless upgrade on Mechain enable a client to sync blocks from genesis to the latest state.

## Quick Start
*Note*: Requires [Go 1.18+](https://go.dev/dl/)

```shell
## proto-all
make proto-all

## build from source
make build

## test
make test

## lint check 
make lint
```

See the [Cosmos Docs](https://cosmos.network/docs/) and [Getting started with the SDK](https://tutorials.cosmos.network/academy/1-what-is-cosmos/).

## Related Projects
- [Mechain](https://github.com/zkMeLabs/mechain): the official Mechain blockchain client.
- [Mechain-Storage-Provider](https://github.com/zkMeLabs/mechain-storage-provider): the storage service infrastructures provided by either organizations or individuals.
- [Mechain-Relayer](https://github.com/zkMeLabs/mechain-relayer): the service that relay cross chain package to both chains.
- [Mechain-Cmd](https://github.com/zkMeLabs/mechain-cmd): the most powerful command line to interact with Mechain system.
- [Awesome Cosmos](https://github.com/cosmos/awesome-cosmos): Collection of Cosmos related resources which also fits Mechain.


## Contribution
Thank you for considering helping with the source code! We appreciate contributions from anyone on the internet, no
matter how small the fix may be.

If you would like to contribute to Mechain, please follow these steps: fork the project, make your changes, commit them,
and send a pull request to the maintainers for review and merge into the main codebase. However, if you plan on submitting
more complex changes, we recommend checking with the core developers first via GitHub issues (we will soon have a Discord channel)
to ensure that your changes align with the project's general philosophy. This can also help reduce the workload of both
parties and streamline the review and merge process.

## Licence

The mechain-cosmos-sdk is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html),
also included in our repository in the `COPYING.LESSER` file.