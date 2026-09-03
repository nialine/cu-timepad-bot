# CU-Timepad-Bot
Telegram-бот для мониторинга появления новых записей (слотов) на мероприятия в Timepad. Бот периодически опрашивает страницу события и присылает уведомление, если появились доступные места.

## Установка и запуск
```
echo 'BOT_TOKEN=YOUR_BOT_TOKEN' > .env
echo 'MONGODB=mongodb://examplelogin:examplepassword@mongo:27017' >> .env

go install
go run cmd/bot
```
