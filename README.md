# Claims fixer

## Requirements

- go
- mainnet genesis (`wget https://archive.evmos.org/mainnet/genesis.json`)
- evmosd node archive 

## Usage

- go run main init
- go run main process

## Results

The file `data_generated.db` has the result of executing the `process` script.
It's a table named `claims` with three columns: 
- `address`
- `balance` 
- `sequence`

NOTE: the accounts with balance `0` are probably accounts that we used to migrate claims/recovery
