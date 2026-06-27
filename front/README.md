# Notifier Admin Panel

A comprehensive admin panel for the Minisource Notifier microservice. Manage notifications, templates, user preferences, and delivery logs across all channels (Email, SMS, Push, In-App).

## Features

### Dashboard
- Overview of notification system metrics
- Quick stats (total sent, failed, pending, today)
- Notifications by type distribution chart
- Quick action buttons for common tasks

### Notifications
- **View all notifications** - Paginated list with search by user ID
- **Unread notifications** - Filter to show unread only
- **Create notification** - Send to any user via any channel
- **Notification detail** - View full details with retry/cancel actions
- **Batch send** - Send multiple notifications at once

### Templates
- **Template management** - CRUD operations for notification templates
- **Template variables** - Support for `{{variable}}` placeholders
- **Provider integration** - Configure provider-specific templates

### Preferences
- Per-user notification channel preferences
- Toggle channels on/off (Email, SMS, Push, In-App)
- Configure instant vs digest delivery
- Set digest frequency (daily, weekly, monthly)

### Admin
- **Users** - Browse users and view their notifications
- **Logs** - View delivery logs and status history
- **Settings** - Configure system-wide settings

## Tech Stack

- **Framework**: Next.js 15 (App Router)
- **Language**: TypeScript
- **Styling**: Tailwind CSS + shadcn/ui
- **State Management**: Zustand
- **Server State**: TanStack React Query
- **Forms**: React Hook Form + Zod
- **Icons**: Lucide React
- **API Client**: Axios

## Getting Started

### Prerequisites

- Node.js 20+
- npm 10+

### Installation

```bash
npm install
```

### Development

```bash
npm run dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

### Build

```bash
npm run build
```

### Docker

```bash
# Development
docker-compose -f docker-compose.dev.yml up

# Production
docker-compose up -d
```

## Environment Variables

Copy `.env.example` to `.env.local` and configure:

```env
NEXT_PUBLIC_APP_NAME=Notifier Admin
NEXT_PUBLIC_APP_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:9002/api
```

## API Endpoints

The panel connects to the Notifier service API:

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health/` | Health check |
| POST | `/api/v1/notifications` | Create notification |
| POST | `/api/v1/notifications/batch` | Batch create |
| GET | `/api/v1/notifications/user/{userId}` | Get user notifications |
| GET | `/api/v1/notifications/user/{userId}/unread` | Get unread |
| PUT | `/api/v1/notifications/{id}/read` | Mark as read |
| GET | `/api/v1/preferences/user/{userId}` | Get preferences |
| PUT | `/api/v1/preferences/user/{userId}` | Update preference |
| GET/POST/PUT/DELETE | `/api/v1/templates` | Template CRUD |
| GET/PUT | `/api/v1/admin/settings` | System settings |

## Project Structure

```
notifier/front/
├── src/
│   ├── api/           # API client and service modules
│   │   └── services/  # Notifications, preferences, templates, admin
│   ├── app/           # Next.js App Router pages
│   │   ├── (auth)/    # Login
│   │   └── (main)/    # Dashboard, notifications, templates, preferences, admin
│   ├── components/    # UI components (shadcn/ui based)
│   │   ├── layout/    # Sidebar, header
│   │   └── ui/        # Button, card, dialog, etc.
│   ├── config/        # Environment config and constants
│   ├── hooks/         # Custom React hooks
│   ├── lib/           # Utility functions
│   ├── stores/        # Zustand stores (auth, UI)
│   ├── styles/        # Global CSS
│   └── types/         # TypeScript type definitions
├── Dockerfile
├── docker-compose.yml
└── README.md
```
