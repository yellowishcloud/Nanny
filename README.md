# 🐾 Nanny — Pet Care Platform

Веб-платформа для поиска и бронирования нянь (pet-sitters) и pet-friendly сервисов в Алматы.  
Помогает владельцам домашних животных быстро и безопасно находить проверенных исполнителей для ухода за питомцами.

---

## 🚀 Цели проекта
- Создать удобное и безопасное онлайн-пространство для владельцев животных.  
- Снизить время и усилия на поиск надёжных ситтеров.  
- Обеспечить прозрачность и доверие при сделках.  

---

## 🔑 Основной функционал
### Для хозяина (Owner):
- Регистрация и авторизация  
- Добавление питомцев  
- Поиск и фильтрация ситтеров (по животному, услуге, локации)  
- Онлайн-бронирование и оплата  
- История заказов  
- Отзывы и рейтинги  

### Для ситтера (Pet-sitter):
- Регистрация и заполнение анкеты (опыт, сертификаты, услуги)  
- Предпочтения (с кем работает / не работает)  
- Интервью и модерация от админа  
- Получение заказов  
- Чат с хозяином, фотоотчёты  
- Получение оплаты через платформу  

### Для админа (Admin):
- Проверка анкет ситтеров  
- Принятие/отклонение заявок на регистрацию ситтеров  
- Модерация профилей и чатов  
- Управление оплатами и спорами  
- Просмотр статистики пользователей и заказов  

---

## 🛠️ Технологии
- **Frontend**: React / Next.js + Tailwind  
- **Backend**: Node.js (Express / NestJS) или Java Spring Boot  
- **Database**: PostgreSQL / MySQL  
- **Auth**: JWT / OAuth2  
- **Payments**: Stripe / KaspiPay  
- **DevOps**: Docker, GitHub Actions  

---

## 📅 Планы разработки
- [x] Анализ конкурентов (Paws.kz)  
- [x] Составление UML и ER диаграмм  
- [ ] Прототип в Figma  
- [ ] MVP: регистрация, профили, поиск ситтеров  
- [ ] Онлайн-оплата и бронирование  
- [ ] Запуск пилота в Алматы  

---

## 👨‍💻 Команда
- Product Manager: …  
- UX/UI Designer: …  
- Backend Developer: …  
- Frontend Developer: …  

---


## 📌 Архитектура

### UML Use Case Diagram
![Use Case Diagram](diagrams/usecase-diagram.png)

### ER Diagram
```mermaid
erDiagram
    USERS {
        int user_id PK
        varchar full_name
        varchar email
        varchar phone
        varchar password_hash
        varchar role  "owner | sitter | admin"
        datetime created_at
    }

    PETS {
        int pet_id PK
        int owner_id FK
        varchar name
        varchar type  "cat | dog | rodent"
        int age
        text notes
    }

    SITTERS {
        int sitter_id PK, FK
        int experience_years
        text certificates
        text preferences
        varchar location
        varchar status  "pending | approved | rejected"
    }

    SERVICES {
        int service_id PK
        int sitter_id FK
        varchar type  "walking | boarding | home-care"
        decimal price_per_hour
        text description
    }

    BOOKINGS {
        int booking_id PK
        int owner_id FK
        int sitter_id FK
        int pet_id FK
        int service_id FK
        datetime start_time
        datetime end_time
        varchar status  "pending | confirmed | cancelled | completed"
    }

    PAYMENTS {
        int payment_id PK
        int booking_id FK
        decimal amount
        varchar method  "card, etc."
        varchar status  "paid | refunded | failed"
        datetime created_at
    }

    REVIEWS {
        int review_id PK
        int booking_id FK
        int owner_id FK
        int sitter_id FK
        int rating
        text comment
        datetime created_at
    }

    CHATS {
        int chat_id PK
        int booking_id FK
        datetime created_at
    }

    MESSAGES {
        int message_id PK
        int chat_id FK
        int sender_id FK
        text content
        datetime sent_at
    }

    USERS ||--o{ PETS : "owns"
    USERS ||--o{ BOOKINGS : "makes"
    USERS ||--o{ SITTERS : "can_be"
    SITTERS ||--o{ SERVICES : "offers"
    SERVICES ||--o{ BOOKINGS : "requested_for"
    BOOKINGS ||--|| PAYMENTS : "has"
    BOOKINGS ||--|| REVIEWS : "receives"
    BOOKINGS ||--|| CHATS : "opens"
    CHATS ||--o{ MESSAGES : "contains"

