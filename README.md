# 🧬 Health Neuro Consultant

A Telegram bot that provides personalized recommendations based on user questionnaire data and genetic analysis results. The bot uses advanced artificial intelligence models to deliver scientifically-backed advice on lifestyle, nutrition, and physical activity.

## ✨ Features

- Personalized recommendations based on user data
- Integration with genetic testing services
- Support for multiple AI models (OpenAI/Anthropic/DeepSeek/Mistral)

## 🛠️ Requirements

- Go 1.24 or higher
- Docker and Docker Compose
- Telegram Bot token
- AI service credentials (OpenAI/Anthropic/DeepSeek/Mistral)
- MyGenetics API credentials (for testing)

## ⚙️ Configuration

1. Copy the example configuration files:
```bash
cp config/example/config.yaml config.yaml
```

2. Update the configuration files with your credentials:
- `config.yaml` - Main application configuration

## 📦 Installation

### Using Docker

```bash
# Build and start services
docker-compose up -d

# View logs
docker-compose logs -f
```

### Manual Installation

```bash
# Install dependencies
go mod download

# Build the application
go build -o health-bot ./cmd/bot

# Run the bot
./health-bot -config config.yaml
```

### 🧪 Running Tests

To run tests, you need to set the following environment variables:

```bash
# For OpenAI tests
export OPENAI_API_KEY=your_api_key
export OPENAI_SOCKS_PROXY=socks5://user:pass@host:port  # optional

# For Anthropic tests
export ANTHROPIC_API_KEY=your_api_key
export ANTHROPIC_SOCKS_PROXY=socks5://user:pass@host:port  # optional

# For DeepSeek tests
export DEEPSEEK_API_KEY=your_api_key
export DEEPSEEK_SOCKS_PROXY=socks5://user:pass@host:port  # optional

# For Mistral tests
export MISTRAL_API_KEY=your_api_key
export MISTRAL_SOCKS_PROXY=socks5://user:pass@host:port  # optional

# For MyGenetics tests
export MYGENETICS_EMAIL=your_email
export MYGENETICS_PASSWORD=your_password
```

After setting the variables, run the tests:
```bash
go test ./...
```

## 👥 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Create a Pull Request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 👨‍💻 Team

- [Valery Polunovsky](https://t.me/vvp310792) - Scientific Lead
- [Olga Shvareva](https://t.me/OlgaShvareva) - NLP Engineer
- [Dmitry Gromazin](https://t.me/Ekzorcist777) - Product Manager
- [Elena Gubskaya](https://t.me/helenatroya729) - ML Engineer
- [Gennadii Muzykantov](https://t.me/muzykantov) - Bioinformatician, Developer

## 📞 Contact

For questions and support, please contact:
- Email: gennadii@muzykantov.me
- Telegram: https://t.me/muzykantov

## 🤖 Bot Link

- [HealthNeuroConsultant](https://t.me/HealthNeuroConsultantBot)