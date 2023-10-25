# File Process
## Overview

The challenge requires creating a system that processes a file from a mounted directory. The file contains a list of debit and credit transactions for an account. The task is to create a function that reads this file, processes its content (where credit transactions are denoted with a plus sign, e.g., +60.5, and debit transactions with a minus sign, e.g., -20.46), and sends a summary of the transactions to a user via email. An example file is provided, but participants are expected to create their own file for the challenge.

## Objectives

- Save the user account data and its transactions.
- The transactions must be processed by reading a CSV file located in a mounted directory.
- The user should receive an email with a summary of their transactions divided by periods.

## Decisions

Opting for a schema where transactions are simply stored in a relational database might seem like the simplest decision, but it is not the correct one from a scalable perspective. Delegating the generation of summaries to the database is a scenario that lacks scalability. Even in the realm of Business Intelligence (BI), thorough data cleaning is essential to ease the process.

In a scenario where all transactions are recorded, the size of the database can grow to a level that incurs costly expenses when running queries. Consequently, even a simple query for a bakery, for instance, becomes a significant challenge in terms of performance and efficiency. It is essential to consider solutions that address this complexity and enable smooth and effective data management.
``
SELECT
s.period,
AVG(s.credit) AS avg_credit,
AVG(s.debit) AS avg_debit,
SUM(a.amount) AS amount
FROM
summaries s
JOIN
accounts a ON s.account_id = a.id
WHERE
s.period = ?
GROUP BY
s.period;
``
This incurs a high computational cost if the table is large.

Given all the aforementioned points, a proper design for this problem should be based on event sourcing, where each event is stored in a database (event store) and projections are generated. For this exercise, there should be two projections: one for the account, where the current account balance is maintained, and another for the summary, where debits and credits are incremented within a given period.

By using an event sourcing schema, the following advantages are gained:


- Decoupling of Components: The update of the account and the summary are not related. Decoupling these different APIs would be straightforward.
- Fault Recovery: In case of a failure, it is possible to reprocess events from a known point.
- Audit Trail: All generated events are stored in the event store. Searching for a specific event would be easy since the account_id is the aggregate_id of the event.

- This project was built using a RESTful API and Event Sourcing architecture, implemented in the Go programming language. The API utilizes Event Sourcing principles with CQRS, where events are stored in a MongoDB event store with two projections in MySQL (account and summary). Additionally, CSV files are saved into a mounted directory managed by the Minio object store. It has an email server where account summaries will be sent. The entire project is containerized using Docker Compose.

## Table of Contents
- [Installation](#installation)
- [Usage](#usage)

## Installation

This project uses Docker Compose for easy setup and deployment. Make sure you have Docker and Docker Compose installed on your system. Follow these steps to get started:
1. Clone the repository:
```bash
   git clone https://github.com/castiglionimax/process-csv.git
   ```

2. Navigate to the project directory:

```bash
cd process-csv
   ```
3. Build and start the project containers:

```bash
docker-compose up --build
   ```
## Usage

To test the application, you can use Postman or a similar tool. Below are the curl commands:

To create a new account:
```sh
curl --location --request POST 'http://127.0.0.1:8080/accounts' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "juan",
    "email": "juan@domain-poc.com"
}'
`````
With the account ID obtained, create a CSV file. There are three ways to do it:

- Using the Minio portal, the username and password are located in the docker-compose file.
URL: http://127.0.0.1:9001/login

    - Username: root
  - Password: Strong#password2023
  
  Go to the object browser -> transactions -> Upload.
  An example CSV file is located in the "csv" folder.

[csv](./csv/)

-Send a POST request and add a JSON payload:


```sh
curl --location --request POST 'http://127.0.0.1:8080/csv' \
--header 'Content-Type: application/json' \
--data-raw '[
  {
    "account_id": {account_id},<---- with the account_id received
    "timestamp": 1634627622,
    "amount": "+9.0"
  },
    {
    "account_id": {account_id},<----with the account_id received
    "timestamp": 1636214203,
    "amount": "-96.5"
  }
]
'
```
Or send multipart/form-data request

```sh
curl --location --request POST 'http://127.0.0.1:8080/csv/upload' \
--form 'csv=@"{LOCATION_FILE}/file.csv"'
```

The cvs must have a following composition:
```sh
account_id,timestamp,amount
bf08ebb5-b470-490e-9b94-192b0e560dd3,1697823898,+60.5
```

To obtain process the files sent.

```sh
curl --location --request POST 'http://127.0.0.1:8080/csv/process'
```

Finally, to get a summary report by email
```sh
curl --location --request POST 'http://127.0.0.1:8080/accounts/ceb7d9ca-36ff-42c7-b394-826498a847f5/summary/email?start=2023-07-01&end=2023-08-01'
```
NOTE: Dates are optional; if not input, the user will receive from the two previous months.

To see the sent email, go to http://127.0.0.1:3000/. This is a fake SMTP server, only for development purposes.

To stop the project containers, you can run:
```bash
docker-compose down
   ```

