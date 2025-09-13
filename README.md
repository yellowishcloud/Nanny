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
