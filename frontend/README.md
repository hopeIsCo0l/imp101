# Frontend - React Authentication App

A simple React frontend for the Imp101 authentication API.

## Features

- User signup with email and password
- User login
- Protected dashboard displaying user information
- JWT token management
- Responsive design

## Setup

1. Install dependencies:
```bash
npm install
```

2. Start the development server:
```bash
npm run dev
```

The app will be available at `http://localhost:3000`

## Environment Variables

Create a `.env` file in the frontend directory to customize the API URL:

```
VITE_API_URL=http://localhost:8080
```

## Build for Production

```bash
npm run build
```

The built files will be in the `dist` directory.
