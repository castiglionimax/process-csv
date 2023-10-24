CREATE TABLE IF NOT EXISTS accounts (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(250) NOT NULL,
    amount DECIMAL(50, 6) NOT NULL,
    last_updated DATETIME NOT NULL
    );


CREATE TABLE IF NOT EXISTS summaries (
    account_id VARCHAR(255) NOT NULL,
    period VARCHAR(255) NOT NULL,
    credit DECIMAL(50, 3) NOT NULL,
    credit_qty INTEGER NOT NULL,
    debit DECIMAL(50, 3) NOT NULL,
    debit_qty INTEGER NOT NULL,
    last_updated DATETIME NOT NULL,
    PRIMARY KEY(account_id, period),
    FOREIGN KEY (account_id) REFERENCES accounts(id)
    );