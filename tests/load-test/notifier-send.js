import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 10 },   // Ramp up to 10 users
    { duration: '1m', target: 30 },    // Stay at 30 users
    { duration: '2m', target: 50 },    // Ramp up to 50 users
    { duration: '1m', target: 50 },    // Stay at 50 users
    { duration: '30s', target: 0 },    // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95% of requests should be below 1s
    http_req_failed: ['rate<0.05'],     // Error rate should be less than 5%
    errors: ['rate<0.05'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:9002';
const API_KEY = __ENV.API_KEY || 'test-api-key';

// Test user IDs
const testUserIds = [
  '00000000-0000-0000-0000-000000000001',
  '00000000-0000-0000-0000-000000000002',
  '00000000-0000-0000-0000-000000000003',
];

export default function () {
  // Select random user
  const userId = testUserIds[Math.floor(Math.random() * testUserIds.length)];

  // Test 1: Send SMS Notification
  const sendSMSPayload = JSON.stringify({
    user_id: userId,
    type: 'sms',
    message: `Load test message ${Date.now()}`,
    metadata: {
      test: true,
      timestamp: Date.now(),
    },
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-API-Key': API_KEY,
    },
    tags: { name: 'SendSMS' },
  };

  const sendSMSRes = http.post(
    `${BASE_URL}/api/v1/notifications/send`,
    sendSMSPayload,
    params
  );

  const sendSMSSuccess = check(sendSMSRes, {
    'SendSMS: status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'SendSMS: response time < 1000ms': (r) => r.timings.duration < 1000,
    'SendSMS: has notification_id': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.notification_id !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!sendSMSSuccess);

  sleep(1);

  // Test 2: Send Email Notification
  const sendEmailPayload = JSON.stringify({
    user_id: userId,
    type: 'email',
    subject: 'Load Test Email',
    message: `Load test email message ${Date.now()}`,
    metadata: {
      test: true,
      timestamp: Date.now(),
    },
  });

  const sendEmailRes = http.post(
    `${BASE_URL}/api/v1/notifications/send`,
    sendEmailPayload,
    params
  );

  const sendEmailSuccess = check(sendEmailRes, {
    'SendEmail: status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'SendEmail: response time < 1000ms': (r) => r.timings.duration < 1000,
  });

  errorRate.add(!sendEmailSuccess);

  sleep(1);

  // Test 3: Get User Notifications
  const getUserNotifRes = http.get(
    `${BASE_URL}/api/v1/notifications/user/${userId}?page=1&limit=10`,
    {
      headers: {
        'X-API-Key': API_KEY,
      },
      tags: { name: 'GetUserNotifications' },
    }
  );

  const getUserNotifSuccess = check(getUserNotifRes, {
    'GetUserNotif: status is 200': (r) => r.status === 200,
    'GetUserNotif: response time < 500ms': (r) => r.timings.duration < 500,
    'GetUserNotif: has notifications array': (r) => {
      try {
        const body = JSON.parse(r.body);
        return Array.isArray(body.notifications);
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!getUserNotifSuccess);

  sleep(2);
}

export function handleSummary(data) {
  return {
    'summary.json': JSON.stringify(data),
    stdout: textSummary(data),
  };
}

function textSummary(data) {
  let summary = '\n';
  summary += 'Notifier Service Load Test Summary:\n';
  summary += '====================================\n';
  summary += `Duration: ${data.state.testRunDurationMs / 1000}s\n`;
  summary += `Iterations: ${data.metrics.iterations.values.count}\n`;
  summary += `VUs: ${data.metrics.vus.values.value}\n`;
  summary += '\n';
  summary += 'HTTP Metrics:\n';
  summary += `  Requests: ${data.metrics.http_reqs.values.count}\n`;
  summary += `  Failed: ${(data.metrics.http_req_failed.values.rate * 100).toFixed(2)}%\n`;
  summary += `  Duration (avg): ${data.metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
  summary += `  Duration (p95): ${data.metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  summary += `  Duration (p99): ${data.metrics.http_req_duration.values['p(99)'].toFixed(2)}ms\n`;
  
  return summary;
}
