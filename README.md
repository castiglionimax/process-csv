# File Process
## Overview

The challenge requires creating a system that processes a file from a mounted directory. The file contains a list of debit and credit transactions for an account. The task is to create a function that reads this file, processes its content (where credit transactions are denoted with a plus sign, e.g., +60.5, and debit transactions with a minus sign, e.g., -20.46), and sends a summary of the transactions to a user via email. An example file is provided, but participants are expected to create their own file for the challenge.

## Objectives
Save the user account data and its transactions.
The transactions must be processed by reading a CSV file located in a mounted directory.
The user should receive an email with a summary of their transactions divided by periods.

## Decisions

Although opting for a schema where transactions are simply stored in a relational database might seem like the simplest decision, it is not the correct one from a scalable perspective. Delegating the generation of summaries to the database is a scenario that lacks scalability. Even in the realm of Business Intelligence (BI), thorough data cleaning is essential to ease the process.

In a scenario where all transactions are recorded, the size of the database can grow to a level that incurs costly expenses when running queries. Consequently, even a simple query for a bakery, for instance, becomes a significant challenge in terms of performance and efficiency. It is essential to consider solutions that address this complexity and enable smooth and effective data management
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
Incurring a high computational cost if the table is large.

Given all the aforementioned points, a proper design for this problem should be based on event sourcing, where each event is stored in a database (event store) and projections are generated:
For this exercise, there should be two projections: one for the account, where the current account balance is maintained, and another for the summary, where debits and credits are incremented within a given period.

By using an event sourcing schema, the following advantages are gained:

- Decoupling of Components: The update of the account and the summary are not related. Decoupling these different APIs would be straightforward.
- Fault Recovery: In case of a failure, it is possible to reprocess events from a known point.
- Audit Trail: All generated events are stored in the event store. Searching for a specific event would be easy since the account_id is the aggregate_id of the event.

This project was built using RestFul api and Event Sourcing architecture, implementing in Go programming language.
The API utilizes Event Sourcing principles con CQRS, where events are stored in a MongoDB event store with two projections in MySQL (account and summary)
Additionally, CSV files are saved into a mount directory managed by  minio object store.
It has an email server where account summaries will be sent. The entire project is containerized using Docker Compose.

## Table of Contents
- [Installation](#installation)
- [Usage](#usage)

## installation

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

To test the application, you can utilize Postman or a similar tool. Below are the curl commands:

To create a new account:
```sh
curl --location --request POST 'http://127.0.0.1:8080/accounts' \
--header 'Content-Type: application/json' \
--data-raw '{
    "name": "juan",
    "email": "juan@domain-poc.com"
}'
`````
With the account ID obtained, make a csv file. There are three form to do it:

- Using minio portal, the user and password are located into the docker-compose file.
url: http://127.0.0.1:9001/login
- usr: root
- pss: Strong#password2023

go to object browser -> transactions -> Upload

a csv example is located into the folder csv.

[csv](./csv/)

- Send a post request and adding a json payload


```sh
curl --location --request POST 'http://127.0.0.1:8080/csv' \
--header 'Content-Type: application/json' \
--data-raw '[
  {
    "account_id": "ceb7d9ca-36ff-42c7-b394-826498a847f5",
    "timestamp": 1634627622,
    "amount": "+9.0"
  },
    {
    "account_id": "ceb7d9ca-36ff-42c7-b394-826498a847f5",
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
NOTE: dates are optional, if those aren't input, the user will receive a summary with two months old. 

In order to see the email sent, go to http://127.0.0.1:3000/ what is a smtp server fake, only for developing propose.

To stop the project containers, you can run:
```bash
docker-compose down
   ```

