# Telegram Nebulas bot https://t.me/nebulasbot

Nebulas telegram bot is an application in the telegram messenger that helps to keep track of transactions and delegations. 
The bot can notify you when the validator node experiences issues and the stability index is reduced. Also you can check balances of your accounts. Here's the full list of features available in the bot. 

Features:
 - Support of the en\cn languages 
 - Add/Remove address for monitoring
 - Add/Remove validator address for monitoring
 - Show the total balance in native tokens and USD. NAS/USD and NAX/USD
 - Show Incoming/Outgoing tx notifications for NAS and NAX
 - Staking/Unstaking of NAX notifications for the validator accounts.
 - Staking/Unstaking of NAS for user accounts
 - Receipt of NAX rewards for NAS staking.
 - Receipt of notifications when the validator's stability index has reduced
 - Add min/max transaction threshold
 - Show the transaction status info 
 - Mute/unmute notifications
 

Dependency:
 - Mysql
 - Nebulas node
 - Golang

## How to run ?
At first you need to configure the config.json file.
```sh
cp config.example.json config.json
```
Next step you need to build and run application.
#### Docker-compose way:
```sh
cp docker-compose.example.yml docker-compose.yml
vim .env // set DB_NAME, DB_USER, DB_PASSWORD for compose file
```
> don`t forget set your passwords
```sh
docker-compose build && docker-compose up -d
```
#### Native way:
> at first setup your dependency and set passwords
```sh
go build && ./nebulas-tg-bot
```
