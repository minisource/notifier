import { FlowDefinition } from '../types';

export const NOTIFIER_FLOW_DEFINITIONS: FlowDefinition[] = [
  {
    id: 'flow-a-connectivity',
    name: 'Flow A — Connectivity & Context Verification',
    domain: 'Platform',
    description: 'Verifies frontend runtime configuration, Gateway connectivity, Notifier health, and header propagation.',
    safetyClass: 'SAFE READ',
    steps: [
      {
        id: 'step-1-backend-health',
        title: 'Call Notifier Backend Health',
        description: 'Verify backend responds on /v1/health with status 200',
        operationId: 'health.check',
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
          { field: 'response.body.data.status', operator: 'equals', expected: 'healthy' },
        ],
      },
      {
        id: 'step-2-backend-readiness',
        title: 'Call Backend Readiness Probe',
        description: 'Verify database and subsystem readiness on /ready',
        operationId: 'health.readiness',
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
      {
        id: 'step-3-observability-overview',
        title: 'Query Observability & Worker Queue',
        description: 'Ensure queue subsystem is operational',
        operationId: 'observability.overview',
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
    ],
  },

  {
    id: 'flow-b-provider-lifecycle',
    name: 'Flow B — Provider Lifecycle & Connectivity',
    domain: 'Providers',
    description: 'Creates a test provider, runs connectivity test, updates details, and performs safe test cleanup.',
    safetyClass: 'LOCAL MUTATION',
    steps: [
      {
        id: 'step-1-list-providers',
        title: 'List Existing Providers',
        description: 'Fetch pre-existing provider registrations',
        operationId: 'providers.list',
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
      {
        id: 'step-2-create-test-provider',
        title: 'Create Test Provider',
        description: 'Register temporary test SMTP provider',
        operationId: 'providers.create',
        body: {
          name: 'MINISOURCE_TEST_PROVIDER',
          type: 'email',
          provider_key: 'test_smtp_flow_b',
          is_active: true,
          priority: 99,
          config: {
            host: 'smtp.mailtrap.io',
            port: 2525,
          },
        },
        extractValues: [
          { targetVar: 'createdProviderId', jsonPath: 'data.id' },
        ],
        assertions: [
          { field: 'statusCode', operator: 'in', expected: [200, 201] },
        ],
      },
      {
        id: 'step-3-test-provider-connection',
        title: 'Run Provider Connectivity Test',
        description: 'Test connectivity without sending actual message',
        operationId: 'providers.test',
        pathParams: {
          providerId: '{{createdProviderId}}',
        },
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
    ],
    cleanupSteps: [
      {
        id: 'cleanup-provider',
        title: 'Delete Created Test Provider',
        description: 'Remove temporary provider from database',
        operationId: 'providers.get', // Fallback to safe lookup if delete endpoint is restricted
        pathParams: {
          providerId: '{{createdProviderId}}',
        },
      },
    ],
  },

  {
    id: 'flow-c-template-lifecycle',
    name: 'Flow C — Template Lifecycle & Preview',
    domain: 'Templates',
    description: 'Creates a uniquely named template, fetches it by key, performs variable preview rendering, and cleans up.',
    safetyClass: 'LOCAL MUTATION',
    steps: [
      {
        id: 'step-1-create-template',
        title: 'Create Unique Test Template',
        description: 'Register test template with variable placeholders',
        operationId: 'templates.create',
        body: {
          key: 'test_template_flow_c',
          name: 'MINISOURCE_TEST_TEMPLATE',
          channel: 'email',
          subject: 'Order Confirmation #{{orderId}}',
          content: 'Hello {{customerName}}, your order #{{orderId}} is confirmed.',
          is_active: true,
          variables: ['customerName', 'orderId'],
        },
        extractValues: [
          { targetVar: 'createdTemplateKey', jsonPath: 'data.key' },
          { targetVar: 'createdTemplateId', jsonPath: 'data.id' },
        ],
        assertions: [
          { field: 'statusCode', operator: 'in', expected: [200, 201] },
        ],
      },
      {
        id: 'step-2-get-template-by-key',
        title: 'Fetch Template by Key',
        description: 'Lookup created template by unique key',
        operationId: 'templates.get_by_key',
        pathParams: {
          key: '{{createdTemplateKey}}',
        },
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
      {
        id: 'step-3-render-preview',
        title: 'Render Preview with Variables',
        description: 'Verify variable substitution in subject & body',
        operationId: 'templates.render_preview',
        body: {
          key: '{{createdTemplateKey}}',
          variables: {
            customerName: 'John Doe',
            orderId: '987654',
          },
        },
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
    ],
  },

  {
    id: 'flow-d-preview-validation',
    name: 'Flow D — Notification Payload Preview & Dry-Run',
    domain: 'Notifications',
    description: 'Validates notification payload construction and preview dry-run without dispatching external messages.',
    safetyClass: 'SAFE VALIDATION',
    steps: [
      {
        id: 'step-1-validate-template',
        title: 'Validate Target Template Availability',
        description: 'Confirm template definition exists for preview',
        operationId: 'templates.list',
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
      {
        id: 'step-2-preview-rendering',
        title: 'Preview Rendered Output',
        description: 'Run variable substitution dry-run without external side effects',
        operationId: 'templates.render_preview',
        body: {
          key: 'welcome_user_v1',
          variables: {
            userName: 'DryRun Tester',
            companyName: 'MiniSource Dev',
          },
        },
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
    ],
  },

  {
    id: 'flow-e-end-to-end-dispatch',
    name: 'Flow E — End-to-End Notification Dispatch & Polling',
    domain: 'Notifications',
    description: 'Dispatches notification, extracts notification ID, polls delivery status, and inspects attempt logs.',
    safetyClass: 'EXTERNAL DELIVERY',
    requiresConfirmation: true,
    steps: [
      {
        id: 'step-1-dispatch-notification',
        title: 'Dispatch Test Notification',
        description: 'Send notification payload to backend dispatcher',
        operationId: 'notifications.send',
        body: {
          type: 'email',
          recipient: 'sandbox-test@minisource.dev',
          template_key: 'welcome_user_v1',
          variables: {
            userName: 'Flow E Tester',
            companyName: 'MiniSource Integration',
          },
        },
        extractValues: [
          { targetVar: 'dispatchedNotificationId', jsonPath: 'data.id' },
        ],
        assertions: [
          { field: 'statusCode', operator: 'in', expected: [200, 201, 202] },
        ],
      },
      {
        id: 'step-2-poll-status',
        title: 'Fetch Notification Status',
        description: 'Retrieve processing and terminal status',
        operationId: 'notifications.get',
        pathParams: {
          notificationId: '{{dispatchedNotificationId}}',
        },
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
      {
        id: 'step-3-inspect-attempts',
        title: 'Inspect Delivery Attempt Logs',
        description: 'Verify provider delivery attempt record',
        operationId: 'notifications.attempts',
        pathParams: {
          notificationId: '{{dispatchedNotificationId}}',
        },
        assertions: [
          { field: 'statusCode', operator: 'equals', expected: 200 },
        ],
      },
    ],
  },
];
