from models import Reading

fake_users_db = {
    "testuser": {
        "username": "testuser",
        "full_name": "Test User",
        "email": "test@example.com",
        # https://cyberchef.org/#recipe=Bcrypt(10)&input=c2VjcmV0
        "hashed_password": "$2a$10$z5scDKQHiJAyXnOOjCgHOulGOVb1I4ehsQT8zw8kz99IJtGYy5u/m",
    }
}

fake_readings_db: dict[str, list[Reading]] = {"testuser": []}
