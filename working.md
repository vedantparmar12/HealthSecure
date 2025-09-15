# HealthSecure Startup Instructions

This document provides the steps to run the HealthSecure application.

## Backend

The backend is a Go application. To run it, navigate to the `backend` directory and run the following commands:

```bash
go mod tidy
go run ./cmd/server
```

The backend server will start on the port specified in the `.env` file (default is 8080).

## Frontend

The frontend is a React application. To run it, navigate to the `frontend` directory and run the following commands:

```bash
npm install
npm start
```

The frontend development server will start, and you can view the application in your browser, usually at `http://localhost:3000`.
