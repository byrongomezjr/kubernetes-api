# Kubernetes API

A production-grade Golang API service designed to demonstrate enterprise-level Kubernetes deployment practices.

## Features

- **Robust Go API**: RESTful API with proper error handling, middleware, and versioning
- **Authentication**: JWT-based authentication system with secure password hashing
- **Database Integration**: PostgreSQL integration with connection pooling and migrations
- **Metrics and Monitoring**: Prometheus metrics endpoint for observability
- **Advanced Kubernetes Deployment**:
  - Resource limits and requests
  - Horizontal Pod Autoscaling
  - Liveness and readiness probes
  - ConfigMaps and Secrets management
  - Network Policies for security
  - Ingress with TLS
  - Pod Disruption Budget for high availability
- **Helm Chart**: Complete Helm chart for easy deployment
- **CI/CD Pipeline**: GitHub Actions workflow for testing, building, and deploying

## Prerequisites

- Go (1.22 or later)
- Docker and Docker Compose
- Kubernetes cluster (for deployment)
- Helm (v3.x)
- kubectl

## Architecture

This application follows a clean, layered architecture:

- `internal/api` - API handlers and routing
- `internal/models` - Data models and DTOs
- `internal/database` - Database connections and data access
- `internal/auth` - Authentication system
- `internal/metrics` - Prometheus metrics
- `pkg/utils` - Common utilities
- `k8s/` - Kubernetes manifests
- `helm/` - Helm chart for Kubernetes deployment

## Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/byrongomezjr/kubernetes-api.git
   cd kubernetes-api
   ```

2. Set up environment:
   ```bash
   cp configs/env.sample .env
   # Edit .env with your configuration
   ```

3. Start dependencies with Docker Compose:
   ```bash
   docker-compose up -d postgres
   ```

4. Build and run the application:
   ```bash
   go build -v ./...
   ./kubernetes-api
   ```

5. Run tests:
   ```bash
   go test -v ./...
   ```

## API Endpoints

### Public Endpoints

- `GET /api/health` - Health check endpoint
- `POST /api/v1/auth/register` - Register a new user
- `POST /api/v1/auth/login` - Authenticate a user

### Protected Endpoints (Require JWT)

- `GET /api/v1/items` - Get all items
- `POST /api/v1/items` - Create a new item
- `GET /api/v1/items/{id}` - Get an item by ID
- `PUT /api/v1/items/{id}` - Update an item
- `DELETE /api/v1/items/{id}` - Delete an item

## Kubernetes Deployment

### Using kubectl

1. Build and push the Docker image:
   ```bash
   docker build -t byrongomezjr/kubernetes-api:latest .
   docker push byrongomezjr/kubernetes-api:latest
   ```

2. Create namespace and deploy:
   ```bash
   kubectl apply -f k8s/namespace.yaml
   kubectl apply -f k8s/secret.yaml
   kubectl apply -f k8s/configmap.yaml
   kubectl apply -f k8s/postgres.yaml
   kubectl apply -f k8s/deployment.yaml
   kubectl apply -f k8s/service.yaml
   kubectl apply -f k8s/ingress.yaml
   kubectl apply -f k8s/hpa.yaml
   kubectl apply -f k8s/pdb.yaml
   kubectl apply -f k8s/networkpolicy.yaml
   ```

### Using Helm

1. Install or upgrade using Helm:
   ```bash
   helm upgrade --install kubernetes-api ./helm/kubernetes-api \
     --namespace kubernetes-api --create-namespace \
     --set image.tag=latest \
     --set secrets.JWT_SECRET="your-secure-secret-key"
   ```

2. Verify the deployment:
   ```bash
   kubectl get pods -n kubernetes-api
   ```

## CI/CD Pipeline

The project uses GitHub Actions for CI/CD. The workflow (`/.github/workflows/ci-cd.yml`) includes:

1. **Build and Test**:
   - Build the application
   - Run unit tests
   - Run linting

2. **Docker Build and Push**:
   - Build the Docker image
   - Push to Docker Hub with appropriate tags

3. **Kubernetes Deployment**:
   - Deploy to Kubernetes using Helm
   - Verify the deployment

To set up the pipeline, you need to configure the following secrets in your GitHub repository:
- `DOCKERHUB_USERNAME` - Docker Hub username
- `DOCKERHUB_TOKEN` - Docker Hub token/password
- `KUBE_CONFIG` - Kubernetes configuration file (base64 encoded)
- `JWT_SECRET` - Secret key for JWT

### Setting up Kubernetes Deployment in CI/CD

For the deployment step to work correctly, you need to configure a valid Kubernetes configuration:

1. **Generate a kubeconfig file** with access to your cluster:
   ```bash
   kubectl config view --minify --flatten > kubeconfig.yaml
   ```

2. **Encode the file to base64**:
   ```bash
   # Linux/macOS
   cat kubeconfig.yaml | base64 -w 0
   
   # Windows (PowerShell)
   [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes((Get-Content -Raw kubeconfig.yaml)))
   ```

3. **Add as a GitHub Secret**:
   - Go to your repository on GitHub
   - Navigate to Settings > Secrets and variables > Actions
   - Click "New repository secret"
   - Name: `KUBE_CONFIG`
   - Value: Paste the base64 encoded string
   
4. **Grant necessary permissions**:
   - Ensure the service account in your kubeconfig has permissions to deploy to the target namespace

**Note**: If you don't have a Kubernetes cluster yet, the deployment step will be skipped automatically.

## Monitoring

The application exposes metrics at the `/metrics` endpoint in Prometheus format. You can configure Prometheus to scrape these metrics and visualize them using Grafana.

## Security Features

- Non-root user in Docker container
- Read-only root filesystem
- Resource limits
- Network policies
- Secure secrets management
- JWT authentication

## Contributing

1. Fork the repository
2. Create a new branch for your changes
3. Make your changes and commit them
4. Push your changes to your fork
5. Create a pull request

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contact

For questions or feedback, please contact me at byrongomezjr@protonmail.com

## Security Notes

### Environment Variables
This application uses environment variables for configuration. Never commit actual secrets to the repository.

1. Copy `.env.template` to `.env` and fill in your values:
   ```bash
   cp .env.template .env
   # Edit .env with your actual values
   ```

2. For Kubernetes deployments, use sealed secrets or a secret management solution:
   ```bash
   # Create your secret from the template
   cp k8s/secret.template.yaml k8s/secret.yaml
   # Edit secret.yaml with your base64 encoded values
   ```

### Kubernetes Secrets
The `k8s/secret.yaml` file is excluded from git. You need to create this file manually based on the template.

## Environment Variables

This application uses the following environment variables for configuration:

### Database Configuration
- `DB_HOST`: PostgreSQL database host (default: `localhost`)
- `DB_PORT`: PostgreSQL database port (default: `5432`)
- `DB_USER`: PostgreSQL database username (default: `postgres`)
- `DB_PASSWORD`: PostgreSQL database password (required)
- `DB_NAME`: PostgreSQL database name (required)
- `DB_SSLMODE`: PostgreSQL SSL mode (default: `disable`, options: `disable`, `require`, `verify-ca`, `verify-full`)

### JWT Configuration
- `JWT_SECRET`: Secret key for JWT token signing (required)

### API Configuration
- `API_PORT`: Port the API server listens on (default: `8080`)
- `API_HOST`: Host the API server binds to (default: `0.0.0.0`)
- `LOG_LEVEL`: Logging level (default: `info`, options: `debug`, `info`, `warn`, `error`)

### Environment
- `ENV`: Application environment (default: `development`, options: `development`, `testing`, `production`)

### Setting Environment Variables

#### Local Development
For local development, copy the example file and set your values:
```bash
cp .env.example .env
# Edit .env with appropriate values
```

#### Docker Compose
When using Docker Compose, environment variables can be set in the `docker-compose.yml` file:
```yaml
services:
  api:
    environment:
      - DB_HOST=postgres
      - DB_PORT=5432
      - DB_USER=postgres
      - DB_PASSWORD=mysecretpassword
      # ... other variables
```

#### Kubernetes
For Kubernetes deployments, sensitive variables should be stored in Secrets:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: kubernetes-api-secrets
  namespace: kubernetes-api
type: Opaque
data:
  DB_PASSWORD: base64_encoded_password
  JWT_SECRET: base64_encoded_jwt_secret
```

Non-sensitive configuration can be stored in ConfigMaps:
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: kubernetes-api-config
  namespace: kubernetes-api
data:
  DB_HOST: postgres.kubernetes-api.svc.cluster.local
  DB_PORT: "5432"
  DB_USER: api_user
  DB_NAME: api_database
  API_PORT: "8080"
  LOG_LEVEL: "info"
  ENV: "production"
```