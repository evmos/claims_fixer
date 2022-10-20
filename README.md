# Claims fixer

## Requirements

- go
- mainnet genesis (`wget https://archive.evmos.org/mainnet/genesis.json`)
- clawback block results (`wget wget https://github.com/v-homsi/clawback/raw/main/evmos_mainnet_5074187_block-results.json`)
- evmosd node archive

## Configuration

- Inside the file `internal/requester.go` set your endpoint. It must be an archive node!

## Usage

Note: if the process is killed nothing will be store in the database.
Note: if you want to run the process again, make sure to remove the `.db` file.

```sh
git clone https://github.com/tharsis/claims_fixer -depth 1
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
```

## Results

The file `accounts_with_claims.db` is generated after running `init`. It parses the genesis file getting all the accounts that had claims records.

The file `results.db` has the result of executing the `process` script.
It's a table named `claims` with the columns:

- `address`
- `amount`
