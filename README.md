# notifier

notifier is a notification service that supports sending notifications via SMS, email, and push notifications.

## Getting Started

### Prerequisites

- Go 1.23.4 or later
- Docker (optional, for containerized deployment)

### Installation

1. Clone the repository:
    ```sh
    git clone https://github.com/minisource/notifier.git
    cd notifier
    ```

2. Install dependencies:
    ```sh
    task install-deps
    ```

### Configuration

Set the application environment to `docker` or `production`. If not set, it defaults to `development`.

Create a [.env](http://_vscodecontentref_/1) file in the root directory and set the necessary environment variables.

### Running the Application

To build and run the application:
```sh
task run
```

