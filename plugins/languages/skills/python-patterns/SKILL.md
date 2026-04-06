---
name: python-patterns
description: >
  Usar cuando se desarrolla con Python: selección de framework web, patrones async,
  type hints, estructura de proyecto, gestión de dependencias, testing con pytest,
  ORM SQLAlchemy, FastAPI, Pydantic y mejores prácticas modernas (Python 3.11+).
  Activador: "proyecto Python", "FastAPI", "Django", "Flask", "async Python",
  "type hints", "pydantic", "pytest", "poetry", "uv", "pyproject.toml".
version: "1.0"
---

# Python — Patrones Modernos (3.11+)

## Selección de Framework

| Caso de Uso | Framework | Por Qué |
|-------------|-----------|---------|
| API REST + validación | **FastAPI** | OpenAPI automático, Pydantic v2, async nativo |
| Web completo + ORM | **Django** | Baterías incluidas, admin, auth |
| Microservicio minimal | **Litestar** | Alternativa moderna a FastAPI, DI integrado |
| Worker / background | **Celery + Redis** | Cola de tareas probada |
| CLI | **Typer** | Click moderno con type hints |
| Script / tooling | **Python puro** | Sin framework si no es necesario |

## Type Hints Modernos

```python
# ✅ Python 3.11+ — usar tipos builtin directamente
def get_users(ids: list[int]) -> dict[str, User]:
    ...

# ✅ Union simplificado
def parse(value: str | int | None) -> str:
    ...

# ✅ TypeAlias para tipos complejos
type UserId = int  # Python 3.12+
# O en 3.11:
from typing import TypeAlias
UserId: TypeAlias = int

# ✅ dataclass moderno con slots
from dataclasses import dataclass

@dataclass(frozen=True, slots=True)
class Point:
    x: float
    y: float

# ✅ Pydantic v2 para validación
from pydantic import BaseModel, field_validator, model_validator

class UserCreate(BaseModel):
    email: str
    age: int

    @field_validator('email')
    @classmethod
    def validate_email(cls, v: str) -> str:
        if '@' not in v:
            raise ValueError('Email inválido')
        return v.lower()
```

## Async Patterns

```python
# ✅ asyncio.gather para operaciones paralelas
import asyncio

async def fetch_user_data(user_id: int) -> UserData:
    user, orders, prefs = await asyncio.gather(
        get_user(user_id),
        get_orders(user_id),
        get_preferences(user_id),
    )
    return UserData(user=user, orders=orders, prefs=prefs)

# ✅ Timeout explícito
async def with_timeout[T](coro: Coroutine[Any, Any, T], seconds: float) -> T:
    async with asyncio.timeout(seconds):  # Python 3.11+
        return await coro

# ✅ AsyncContextManager para recursos
from contextlib import asynccontextmanager

@asynccontextmanager
async def db_transaction(db: AsyncSession):
    async with db.begin():
        try:
            yield db
        except Exception:
            await db.rollback()
            raise
```

## Gestión de Dependencias

```toml
# pyproject.toml (estándar moderno — usar uv o poetry)
[project]
name = "mi-app"
version = "0.1.0"
requires-python = ">=3.11"
dependencies = [
    "fastapi>=0.115",
    "pydantic>=2.0",
    "sqlalchemy[asyncio]>=2.0",
]

[project.optional-dependencies]
dev = ["pytest>=8", "pytest-asyncio", "httpx", "ruff", "mypy"]

[tool.ruff]
line-length = 100
target-version = "py311"

[tool.mypy]
strict = true
python_version = "3.11"
```

```bash
# uv — gestión de entornos moderna (más rápida que pip)
uv venv && uv pip install -e ".[dev]"
# O poetry:
poetry install --with dev
```

## Estructura de Proyecto

```
src/
└── mi_app/
    ├── __init__.py
    ├── main.py              # Entry point FastAPI
    ├── config.py            # Settings con Pydantic BaseSettings
    ├── database.py          # Engine y sesiones SQLAlchemy
    ├── models/              # Modelos ORM (SQLAlchemy)
    ├── schemas/             # Schemas Pydantic (entrada/salida)
    ├── routers/             # Routers FastAPI por dominio
    ├── services/            # Lógica de negocio
    ├── repositories/        # Acceso a datos
    └── dependencies.py      # Dependencias FastAPI (DI)
tests/
├── conftest.py
├── unit/
└── integration/
```

## Gestión de Errores

```python
# ✅ Excepciones de dominio tipadas
class AppError(Exception):
    def __init__(self, message: str, code: str) -> None:
        super().__init__(message)
        self.code = code

class NotFoundError(AppError):
    def __init__(self, resource: str, id: int | str) -> None:
        super().__init__(f"{resource} {id} not found", "NOT_FOUND")

# ✅ Result type para errores esperados
from dataclasses import dataclass
from typing import Generic, TypeVar

T = TypeVar('T')
E = TypeVar('E', bound=Exception)

@dataclass
class Ok(Generic[T]):
    value: T

@dataclass
class Err(Generic[E]):
    error: E

Result = Ok[T] | Err[E]
```

## Testing con pytest

```python
# conftest.py
import pytest
from httpx import AsyncClient, ASGITransport
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession

@pytest.fixture
async def db_session():
    engine = create_async_engine("sqlite+aiosqlite:///:memory:")
    async with AsyncSession(engine) as session:
        yield session

@pytest.fixture
async def client(app):
    async with AsyncClient(
        transport=ASGITransport(app=app),
        base_url="http://test"
    ) as c:
        yield c

# test_users.py
async def test_create_user(client: AsyncClient):
    response = await client.post("/users", json={"email": "test@example.com", "age": 25})
    assert response.status_code == 201
    assert response.json()["email"] == "test@example.com"
```

## Configuración con Pydantic Settings

```python
from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", extra="ignore")

    debug: bool = False
    database_url: str
    secret_key: str
    allowed_origins: list[str] = ["http://localhost:3000"]

settings = Settings()  # Falla en arranque si falta variable requerida
```
