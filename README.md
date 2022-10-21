# Claims fixer

## Requirements

- go
- mainnet genesis (`wget https://archive.evmos.org/mainnet/genesis.json`)
- clawback block results (`wget https://github.com/v-homsi/clawback/raw/main/evmos_mainnet_5074187_block-results.json`)
- evmosd node archive

## Configuration

- Inside the file `internal/requester.go` set your endpoint. It must be an archive node!

## Usage

Note: if the process is killed nothing will be store in the database.
Note: if you want to run the process again, make sure to remove the `.db` file.

```sh
git clone https://github.com/evmos/claims_fixer --depth 1
cd claims_fixer
wget https://archive.evmos.org/mainnet/genesis.json
wget https://github.com/v-homsi/clawback/raw/main/evmos_mainnet_5074187_block-results.json
go build
rm accounts_with_claims.db
rm results.db
./claims_fixer init
./claims_fixer process
./claims_fixer addAttestationRecords
./claims_fixer processAttestationRecords
./claims_fixer create
```

## Results

The file `results.db` has the result of executing the `process` script.
It's a table named `claims` with the columns:

- `address`
- `amount`

## How it works

- `init`: gets all the wallets that had claims records on the genesis file.
- `process`: check if the account was affected by the clawback bug and add that account and the amount to the `results.db` database.
- `addAttestationRecords`: using the `block_result` information from the clawback block, we all the addresses that sent their coins to the community wallet. Note: this block didn't have any transaction, so most of the wallets inside this file are from the clawback function.
- `processAttestationRecords`: iterates using all the wallets added in the previous step and it makes sure that their balance was incorrectly moved.
- `create`: it generates an `account.txt` file that contains the information needed for the evmosd network upgrade.
