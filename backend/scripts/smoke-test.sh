#!/bin/bash
# ============================================
# Notifier Service — Smoke Test Script
# ============================================
# Usage:
#   BASE_URL=http://localhost:9002 ./scripts/smoke-test.sh
#   BASE_URL=http://localhost:9002 USER_TOKEN=xxx ./scripts/smoke-test.sh
#   BASE_URL=http://localhost:9002 ADMIN_TOKEN=xxx ./scripts/smoke-test.sh
# ============================================

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:9002}"
USER_TOKEN="${USER_TOKEN:-}"
ADMIN_TOKEN="${ADMIN_TOKEN:-}"
SERVICE_TOKEN="${SERVICE_TOKEN:-}"

PASS=0
FAIL=0

print_result() {
    local name=$1
    local status=$2
    if [ "$status" == "pass" ]; then
        echo "  ✅ PASS: $name"
        PASS=$((PASS + 1))
    else
        echo "  ❌ FAIL: $name"
        FAIL=$((FAIL + 1))
    fi
}

echo "============================================"
echo " Notifier Service — Smoke Test"
echo " URL: $BASE_URL"
echo "============================================"
echo ""

# ----- 1. Public Health -----
echo "--- 1. Public Health ---"

HEALTH=$(curl -sf "$BASE_URL/api/v1/health/" 2>&1 || echo "")
if echo "$HEALTH" | grep -q "ok\|healthy\|status"; then
    print_result "GET /api/v1/health/" pass
else
    print_result "GET /api/v1/health/" fail
fi

# ----- 2. Swagger -----
echo "--- 2. Swagger ---"

SWAGGER=$(curl -sf "$BASE_URL/swagger/index.html" 2>&1 || echo "")
if echo "$SWAGGER" | grep -q "swagger\|html\|Swagger"; then
    print_result "GET /swagger/index.html" pass
else
    print_result "GET /swagger/index.html" fail
fi

# ----- 3. Admin Dashboard (no token) -----
echo "--- 3. Admin Access (no token) ---"

ADMIN_NO_TOKEN=$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/api/v1/admin/dashboard/overview" 2>&1 || echo "")
if [ "$ADMIN_NO_TOKEN" == "401" ] || [ "$ADMIN_NO_TOKEN" == "403" ]; then
    print_result "Admin dashboard without token → 401/403" pass
else
    print_result "Admin dashboard without token → $ADMIN_NO_TOKEN (expected 401/403)" fail
fi

# ----- 4. Admin Dashboard (with admin token) -----
echo "--- 4. Admin Dashboard (with token) ---"

if [ -n "$ADMIN_TOKEN" ]; then
    ADMIN_OK=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "$BASE_URL/api/v1/admin/dashboard/overview" 2>&1 || echo "")
    if [ "$ADMIN_OK" == "200" ]; then
        print_result "Admin dashboard with admin token → 200" pass
    else
        print_result "Admin dashboard with admin token → $ADMIN_OK (expected 200)" fail
    fi
else
    echo "  ⚠️  ADMIN_TOKEN not set, skipping admin dashboard test"
fi

# ----- 5. Admin Observability (with admin token) -----
echo "--- 5. Admin Observability ---"

if [ -n "$ADMIN_TOKEN" ]; then
    METRICS=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "$BASE_URL/api/v1/admin/observability/metrics" 2>&1 || echo "")
    if [ "$METRICS" == "200" ]; then
        print_result "Admin metrics → 200" pass
    else
        print_result "Admin metrics → $METRICS (expected 200)" fail
    fi

    QUEUE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "$BASE_URL/api/v1/admin/observability/queue" 2>&1 || echo "")
    if [ "$QUEUE" == "200" ]; then
        print_result "Admin queue overview → 200" pass
    else
        print_result "Admin queue overview → $QUEUE (expected 200)" fail
    fi

    HEALTH_ADMIN=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $ADMIN_TOKEN" \
        "$BASE_URL/api/v1/admin/observability/health" 2>&1 || echo "")
    if [ "$HEALTH_ADMIN" == "200" ]; then
        print_result "Admin health → 200" pass
    else
        print_result "Admin health → $HEALTH_ADMIN (expected 200)" fail
    fi
else
    echo "  ⚠️  ADMIN_TOKEN not set, skipping admin observability tests"
fi

# ----- 6. User /me (with user token) -----
echo "--- 6. User /me Endpoints ---"

if [ -n "$USER_TOKEN" ]; then
    ME_NOTIF=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $USER_TOKEN" \
        "$BASE_URL/api/v1/me/notifications" 2>&1 || echo "")
    if [ "$ME_NOTIF" == "200" ]; then
        print_result "GET /me/notifications → 200" pass
    else
        print_result "GET /me/notifications → $ME_NOTIF (expected 200)" fail
    fi

    ME_PREFS=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $USER_TOKEN" \
        "$BASE_URL/api/v1/me/preferences" 2>&1 || echo "")
    if [ "$ME_PREFS" == "200" ]; then
        print_result "GET /me/preferences → 200" pass
    else
        print_result "GET /me/preferences → $ME_PREFS (expected 200)" fail
    fi

    # Normal user accessing admin should fail
    USER_ADMIN=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $USER_TOKEN" \
        "$BASE_URL/api/v1/admin/dashboard/overview" 2>&1 || echo "")
    if [ "$USER_ADMIN" == "403" ]; then
        print_result "Normal user cannot access admin → 403" pass
    else
        print_result "Normal user accessing admin → $USER_ADMIN (expected 403)" fail
    fi
else
    echo "  ⚠️  USER_TOKEN not set, skipping user /me tests"
fi

# ----- 7. Service Creation (with service token) -----
echo "--- 7. Service Notification Creation ---"

if [ -n "$SERVICE_TOKEN" ]; then
    CREATE=$(curl -s -o /dev/null -w "%{http_code}" \
        -H "Authorization: Bearer $SERVICE_TOKEN" \
        -H "Content-Type: application/json" \
        -X POST "$BASE_URL/api/v1/service/notifications" \
        -d '{
            "userId": "00000000-0000-0000-0000-000000000001",
            "type": "in_app",
            "body": "Smoke test notification"
        }' 2>&1 || echo "")
    if [ "$CREATE" == "201" ] || [ "$CREATE" == "200" ]; then
        print_result "Service notification create → $CREATE" pass
    else
        print_result "Service notification create → $CREATE (expected 201/200)" fail
    fi
else
    echo "  ⚠️  SERVICE_TOKEN not set, skipping service creation test"
fi

# ----- Summary -----
echo ""
echo "============================================"
echo " Results: $PASS passed, $FAIL failed"
echo "============================================"

if [ "$FAIL" -gt 0 ]; then
    exit 1
fi
exit 0
