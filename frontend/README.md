# FlowRack WMS Frontend

Warehouse Management System frontend built with Next.js App Router and integrated to your backend REST API.

## Stack

- Zod for request/response validation
- Zustand for UI state management
- TanStack React Query for data fetching and cache invalidation
- Tailwind CSS for styling
- Shadcn-style UI primitives (`components/ui`) and Headless UI dialog
- TanStack Table for order table rendering
- React Hook Form + Zod resolver for shipping form

## Implemented Backend Integration

WMS endpoints:

- `GET /warehouse/orders?wms_status=...`
- `GET /warehouse/orders/:order_sn`
- `POST /warehouse/orders/:order_sn/pick`
- `POST /warehouse/orders/:order_sn/pack`
- `POST /warehouse/orders/:order_sn/ship`

Marketplace logistics support endpoint:

- `GET /logistic/channels`

## Local Setup

1. Install dependencies

```bash
npm install
```

2. Configure API base URL

```bash
cp .env.example .env.local
```

3. Start development server

```bash
npm run dev
```

4. Open application

`http://localhost:3000`

## Production Validation

```bash
npm run lint
npm run build
```
