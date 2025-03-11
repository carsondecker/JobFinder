# JobFinder

## Getting Started
- Install Go from [here](https://golang.org/dl/)
- Install PostgreSQL from [here](https://www.postgresql.org/download/)
- Download the code in this repository

## Project Setup
#### Set up database
```
# Create database
psql -U postgres -c "CREATE DATABASE projectdb;"

# Run migrations with Goose
goose -dir ./sql/schema postgres "user=postgres dbname=projectdb sslmode=disable" up
```

#### Configure .env file
The .env file should look something like this:
```
DB_URL="{database connection string for postgres, looks like "postgres://postgres:postgres@localhost:5432/jobfinder"}"
PLATFORM="dev" (or anything else for production)
SECRET="{random JWT secret, just generate a random string here}"
```

# JobFinder API Documentation

This document outlines the available endpoints, request formats, and expected responses for the JobFinder API.

## Authentication
Most endpoints require authentication via a Bearer token in the `Authorization` header.

### Register
**POST** `/api/auth/register`

**Request Body:**
```json
{
    "username": "string",
    "password": "string"
}
```

**Response:**
```json
{
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "username": "string"
}
```

### Login
**POST** `/api/auth/login`

**Request Body:**
```json
{
    "username": "string",
    "password": "string"
}
```

**Response:**
```json
{
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "username": "string",
    "token": "jwt_token"
}
```

## Jobs

### Create Job
**POST** `/api/jobs`

**Request Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
    "title": "string",
    "description": "string",
    "city": "string"
}
```

**Response:**
```json
{
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "title": "string",
    "description": "string",
    "city": "string",
    "user_id": "uuid"
}
```

### Get All Jobs
**GET** `/api/jobs`

**Query Parameters (Optional):** `?title=<string>`

**Response:**
```json
[
    {
        "id": "uuid",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "title": "string",
        "description": "string",
        "city": "string",
        "user_id": "uuid"
    }
]
```

### Get Job by ID
**GET** `/api/jobs/{jobID}`

**Response:**
```json
{
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "title": "string",
    "description": "string",
    "city": "string",
    "user_id": "uuid"
}
```

### Delete Job
**DELETE** `/api/jobs/{jobID}`

**Request Headers:**
```
Authorization: Bearer {token}
```

**Response:** 200 OK

## Applications

### Apply for a Job
**POST** `/api/jobs/{jobID}/apply`

**Request Headers:**
```
Authorization: Bearer {token}
```

**Request Body:**
```json
{
    "cover_note": "string"
}
```

**Response:**
```json
{
    "id": "uuid",
    "created_at": "timestamp",
    "updated_at": "timestamp",
    "cover_note": "string",
    "job_id": "uuid",
    "user_id": "uuid"
}
```

### Get Applications by User
**GET** `/api/application`

**Request Headers:**
```
Authorization: Bearer {token}
```

**Response:**
```json
[
    {
        "id": "uuid",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "cover_note": "string",
        "job_id": "uuid",
        "user_id": "uuid"
    }
]
```

### Get Applications for a Specific Job
**GET** `/api/jobs/{jobID}/applications`

**Request Headers:**
```
Authorization: Bearer {token}
```

**Response:**
```json
[
    {
        "id": "uuid",
        "created_at": "timestamp",
        "updated_at": "timestamp",
        "cover_note": "string",
        "job_id": "uuid",
        "user_id": "uuid"
    }
]
```

## Admin Operations

### Reset Database (Development Only)
**POST** `/admin/reset`

**Response:** 200 OK (Only available in `dev` mode)

---

## Error Responses

**General Error Response:**
```json
{
    "error": "error message"
}
```

