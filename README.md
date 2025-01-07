# Ollama DNS Router

This repository provides a Go application to monitor the health of Ollama instances and dynamically update Pi-hole DNS records to route traffic to healthy instances. Designed for DevOps and developers, the tool ensures high availability and seamless instance failover by programmatically managing DNS entries. 

**Currently only Pihole is supported and is in super early stage, might add new providers in the future.**

---

## Features

- **Ollama Instance Monitoring**: Performs HTTP health checks on a list of Ollama instances.
- **Dynamic DNS Management**: Updates Pi-hole DNS records to point to a healthy Ollama instance.
- **Structured Logging**: Uses Zap for detailed, structured logs for operational clarity.
- **Environment-Driven Configuration**: Fully configurable through environment variables for easy CI/CD integration.

---

## Configuration

The application uses environment variables to define its behavior. Below is a detailed table of supported configurations:

| Environment Variable  | Required | Description                                         | Example                           |
|-----------------------|----------|-----------------------------------------------------|-----------------------------------|
| `INSTANCE_URLS`       | ✅       | Comma-separated list of Ollama instance URLs to monitor. | `http://ollama1.local,http://ollama2.local` |
| `OLLAMA_RECORD`       | ✅       | The DNS record to update in Pi-hole.                | `ollama.mydomain.local`           |
| `PIHOLE_HOSTNAME`     | ✅       | The hostname or IP address of the Pi-hole server.   | `192.168.1.100`                   |
| `PIHOLE_API_KEY`      | ✅       | The API key for authenticating with Pi-hole.        | `your_api_key_here`               |

---

## How It Works

1. **Instance Health Check**:
   - The application iterates through the list of Ollama instances provided in `INSTANCE_URLS`.
   - Sends an HTTP `GET` request to each instance.
   - Identifies the first healthy instance (responding with a `200 OK` status).

2. **Dynamic DNS Record Update**:
   - Resolves the IP address of the healthy instance.
   - Deletes any existing DNS record for `OLLAMA_RECORD` in Pi-hole.
   - Creates a new DNS record in Pi-hole that points to the healthy instance's IP address.

3. **Logging**:
   - Outputs structured logs for operational transparency, including instance health and DNS management outcomes.

---

## Setup

### Prerequisites

- **Pi-hole Server**: Ensure Pi-hole is running and its API is enabled.
- **API Key**: Retrieve the Pi-hole API key from the Pi-hole admin interface.
- **Go Environment**: Install Go 1.18+ on your system.

### Installation

1. Clone the repository:
   ```bash
   git clone <repository-url>
   cd <repository-name>
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Build the application:
   ```bash
   go build -o ollama-health-checker
   ```

### Configuration

Set the required environment variables in your terminal or through a `.env` file:

```bash
export INSTANCE_URLS=http://ollama1.local,http://ollama2.local
export OLLAMA_RECORD=ollama.mydomain.local
export PIHOLE_HOSTNAME=192.168.1.100
export PIHOLE_API_KEY=your_api_key_here
```

Alternatively, create a `.env` file:

```env
INSTANCE_URLS=http://ollama1.local,http://ollama2.local
OLLAMA_RECORD=ollama.mydomain.local
PIHOLE_HOSTNAME=192.168.1.100
PIHOLE_API_KEY=your_api_key_here
```

---

## Usage

Run the application:

```bash
./ollama-health-checker
```

Upon execution:

1. The application checks the health of Ollama instances in the order listed in `INSTANCE_URLS`.
2. Updates the DNS record in Pi-hole to route traffic to the first healthy instance.
3. Logs the results of health checks and DNS updates.

---

## Logging

Logs are generated in a structured JSON format for easy integration with log management tools like ELK or Loki.

**Example Log Output**:

```json
{
  "level": "info",
  "msg": "DNS record created successfully",
  "ip": "192.168.1.101",
  "domain": "ollama.mydomain.local"
}
```

**Log Levels**:
- `INFO`: Successful operations (e.g., DNS record creation).
- `ERROR`: Failures (e.g., instance health check failure or DNS update failure).

---

## Troubleshooting

| Issue                           | Possible Cause                                  | Resolution                                                           |
|---------------------------------|------------------------------------------------|----------------------------------------------------------------------|
| **No healthy instances found**  | All instances are down or unreachable.          | Verify instance URLs and ensure they respond with `200 OK`.          |
| **Failed to create DNS record** | Incorrect Pi-hole hostname or API key.          | Double-check `PIHOLE_HOSTNAME` and `PIHOLE_API_KEY`.                 |
| **Timeout during health check** | Network issues or instance unavailability.      | Increase the HTTP client timeout in the code if necessary.           |

---

## Future Enhancements

- Add support for other DNS providers (e.g., AWS Route 53, Cloudflare).
- Implement retries and exponential backoff for instance health checks.
- Integrate metrics collection for long-term monitoring and analysis.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Contributing

Pull requests and issues are welcome! Follow these steps to contribute:

1. Fork the repository.
2. Create a feature branch.
3. Commit your changes and open a pull request.

