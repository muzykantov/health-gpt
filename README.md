# 🧬 Нейроконсультант по здоровью

Телеграм-бот для предоставления персонализированных рекомендаций, основанных на анкетных данных пользователей и результатах их генетического анализа. Бот использует передовые модели искусственного интеллекта для предоставления научно обоснованных советов по образу жизни, питанию и физической активности.

## ✨ Возможности

- Персонализированные рекомендации на основе данных пользователя
- Интеграция с сервисами генетического тестирования
- Поддержка нескольких моделей ИИ (GigaChat, ChatGPT)

## 🛠️ Необходимое ПО

- Go 1.23 или выше
- Docker и Docker Compose
- Токен Telegram Bot
- Учетные данные сервисов ИИ (GigaChat/ChatGPT)
- Учетные данные API MyGenetics (для тестов)

## ⚙️ Конфигурация

1. Скопируйте примеры конфигурационных файлов:
```bash
cp config/example/config.json config.json
cp config/example/.env .env
```

2. Обновите конфигурационные файлы своими учетными данными:
- `config.json` - Основная конфигурация приложения
- `.env` - Переменные окружения

## 📦 Установка

### Использование Docker

```bash
# Собрать и запустить сервисы
docker-compose up -d

# Просмотр логов
docker-compose logs -f
```

### Ручная установка

```bash
# Установка зависимостей
go mod download

# Сборка приложения
go build -o health-bot ./cmd/bot

# Запуск бота
./health-bot -config config.json
```

### 🧪 Запуск тестов

Для запуска тестов необходимо установить следующие переменные окружения:

```bash
# Для тестов ChatGPT
export OPENAI_API_KEY=your_api_key
export OPENAI_SOCKS_PROXY=socks5://user:pass@host:port  # опционально

# Для тестов GigaChat
export GIGACHAT_CLIENT_ID=your_client_id
export GIGACHAT_CLIENT_SECRET=your_client_secret

# Для тестов MyGenetics
export MYGENETICS_EMAIL=your_email
export MYGENETICS_PASSWORD=your_password
```

После установки переменных запустите тесты:
```bash
go test ./...
```

## 👥 Участие в разработке

1. Сделайте форк репозитория
2. Создайте ветку для функционала (`git checkout -b feature/amazing-feature`)
3. Зафиксируйте изменения (`git commit -m 'Add amazing feature'`)
4. Отправьте ветку в удаленный репозиторий (`git push origin feature/amazing-feature`)
5. Создайте Pull Request

## 📄 Лицензия

Этот проект распространяется под лицензией MIT - подробности см. в файле LICENSE.

## 👨‍💻 Команда

- [Ольга Шварева](https://t.me/OlgaShvareva) - Инженер по обработке естественного языка (NLP Engineer)
- [Дмитрий Громазин](https://t.me/Ekzorcist777) - Менеджер продукта
- [Елена Губская](https://t.me/helenatroya729) - Инженер по машинному обучению (ML Engineer)
- [Геннадий Музыкантов](https://t.me/muzykantov) - Биоинформатик, разработчик

## 📞 Контакты

По вопросам и поддержке обращайтесь:
- Email: gennadii@muzykantov.me
- Telegram: https://t.me/muzykantov

## 🤖 Ссылка на бота

- [HealthNeuroConsultant](https://t.me/HealthNeuroConsultantBot)
