# syntax=docker/dockerfile:1

# ── Builder ──────────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /src

ENV NEXT_TELEMETRY_DISABLED=1

# NOTE: design-system is not yet published to npm (Phase 6 pending).
# Until then, this Dockerfile requires a monorepo context that includes
# both design-system/ and notifier/front/ (i.e. build from MiniSource/ root).
# After Phase 6: replace file: protocol with npm-published @minisource/* packages.

# Copy design-system (needs to be in build context)
COPY design-system/ ./design-system/
COPY notifier/front/package.json notifier/front/package-lock.json ./notifier/front/

# Install and build design-system (pnpm workspaces)
WORKDIR /src/design-system
RUN corepack enable && corepack prepare pnpm@latest --activate
RUN pnpm install --no-frozen-lockfile
RUN pnpm build

# Install notifier/front dependencies
WORKDIR /src/notifier/front
RUN npm ci --legacy-peer-deps --no-audit --no-fund

# Build Next.js (standalone output)
ENV NEXT_PUBLIC_NOTIFIER_AUTH_ENABLED=true
RUN npm run build

# ── Runtime ──────────────────────────────────────────────────────────────────
FROM node:20-alpine AS runner

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1

RUN addgroup --system --gid 1001 nodejs && \
    adduser --system --uid 1001 nextjs

COPY --from=builder /src/notifier/front/public ./public
COPY --from=builder /src/notifier/front/.next/standalone ./
COPY --from=builder /src/notifier/front/.next/static ./.next/static

RUN chown -R nextjs:nodejs /app
USER nextjs

WORKDIR /app
EXPOSE 3004

ENV PORT=3004
ENV HOSTNAME=0.0.0.0

HEALTHCHECK --interval=30s --timeout=30s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://127.0.0.1:3004/notifier/dashboard || exit 1

CMD ["node", "server.js"]
