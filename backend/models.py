import datetime
from pydantic import BaseModel


class Token(BaseModel):
    access_token: str
    token_type: str


class Reading(BaseModel):
    timestamp: datetime.datetime
    value: float


class ReadingResponse(BaseModel):
    username: str
    data: list[Reading]


class User(BaseModel):
    username: str
    email: str | None = None
    full_name: str | None = None


class DBUser(User):
    hashed_password: str
