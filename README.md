# Wallet Service

This is a simple wallet service implemented in Go using the Gin framework and a SQL database.

## Endpoints

- `PUT /api/balance/:id` - Deposit or withdraw funds from a wallet.
- `GET /api/balance/:id` - Get the balance of a wallet.
- `POST /api/transfer` - Transfer funds between wallets.
- `GET /api/transaction/:id` - Get the transactions of a wallet.

```
curl 127.0.0.1:8080/api/balance/1
curl 127.0.0.1:8080/api/balance/2
curl -XPUT 127.0.0.1:8080/api/balance/1 -d '{"op_type":"deposit","amount":60}'
curl -XPUT 127.0.0.1:8080/api/balance/1 -d '{"op_type":"withdraw","amount":30}'
curl -XPOST 127.0.0.1:8080/api/transfer -d '{"from_wallet_id":2,"to_wallet_id":1,"amount":20}'
```

## Environment Variables

- `DB_USERNAME` - The username for the database.
- `DB_PASSWORD` - The password for the database.
- `DB_NAME` - The name of the database.
- `DB_HOST` - The host of the database.

## Running the Service

1. Run the service:
```
docker build -t wallet .
docker-compose up -d
```

