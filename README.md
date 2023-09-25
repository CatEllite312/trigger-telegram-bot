```markdown
# Telegram Розважальний Бот

Цей репозиторій містить вихідний код Telegram-бота, який має виключно розважальну функцію. Бот дозволяє зберігати текстові та медіа повідомлення, які бот відправляє повторно, коли в чаті з'являються слова або словосполучення тригери.

## Функції

- Зберігання повідомлень як тригерів для автоматичного пересилання.
- Підтримка різних типів медіа, включаючи текст, фотографії, стікери, відео, голосові повідомлення, аудіо, анімації та відеозаписи.

## Використання

1. **Клонування Репозиторію:**

   ```shell
   git clone https://github.com/your-username/telegram-entertainment-bot.git
   cd telegram-entertainment-bot
   ```

2. **Налаштування Конфігурації:**

   - Створіть файл `config.txt` у кореневій папці проекту.
   - Додайте свій API ключ Telegram Bot до файлу `config.txt`. Якщо файл не існує, бот запросить вас ввести ключ при першому запуску.

3. **Запуск Бота:**

   - Переконайтеся, що на вашій системі встановлено Go.
   - Запустіть бота за допомогою наступної команди:

     ```shell
     go run main.go
     ```

4. **Взаємодія з Ботом:**

   - У чаті в Telegram використовуйте наступні команди для взаємодії з ботом:
     - `/start`: Початок роботи з ботом.
     - `/save <тригер>`: Зберегти тригер для повідомлень (повинно бути використано як відповідь на повідомлення, яке ви хочете зберегти).
     - `/del <тригер>`: Видалити збережений тригер.
     - `/list`: Переглянути список збережених тригерів.
     - `/ping`: Перевірити доступність бота та час відповіді.
     - `/help`: Отримати список доступних команд.

5. **Насолоджуйтеся!**

## Підтримувані Операційні Системи

- Linux
- macOS
- Windows

## Ліцензія

Цей проект розповсюджується за ліцензією MIT
```
