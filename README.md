# Kubernetes API

A Go-based API service designed to interact with Kubernetes clusters.

## Project Structure
.
├── .github/workflows
│ └── ci-cd.yml
├── k8s/
│ └── deployment.yaml
├── Dockerfile
├── go.mod
├── main.go
├── LICENSE
└── README.md

## Prerequisites
- Go (1.22 or later)
- Docker
- Kubernetes cluster (for deployment)

## API Endpoints
- `GET /api/health` - Health check endpoint
- `GET /api/data` - Sample data endpoint

## Local Development
1. Clone the repository:
   ```bash
   git clone https://github.com/byrongomezjr/kubernetes-api.git
   cd kubernetes-api
   ```
2. Build and run the application:
   ```bash
   go build -v ./...
   go run main.go
   ```
3. Test the API:
   ```bash
   go test -v ./...
   ```
4. Test endpoints locally:
   ```bash
   curl http://localhost:8080/api/health
   curl http://localhost:8080/api/data
   ```

## Deployment
1. Build the Docker image:
   ```bash
   docker build -t byrongomezjr/kubernetes-api:latest .
   ```
2. Push the image to Docker Hub:
   ```bash
   docker push byrongomezjr/kubernetes-api:latest
   ```
3. Apply the Kubernetes deployment:
   ```bash
   kubectl apply -f k8s/deployment.yaml
   ```

## CI/CD Pipeline
The project uses GitHub Actions for CI/CD. The workflow is defined in `.github/workflows/ci-cd.yml` and includes:
- Building the application
- Running tests
- Building and pushing Docker image
- (Future: Kubernetes deployment)

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

