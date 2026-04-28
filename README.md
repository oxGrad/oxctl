# oxctl

An opinionated AWS CLI wrapper for deploying ECS Fargate services. oxctl automates the tedious steps of updating container images, registering task definitions, and managing service state, so you can deploy in one command.

**Module:** `github.com/oxGrad/oxctl`

---

## Quick Start

### Install

```bash
# Build from source
go build -o oxctl ./cmd/oxctl

# Or use make
make build
```

### First Deploy

Create an `oxconf` file in your repo root:

```yaml
cluster: my-cluster
service: my-service
image: 123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:abc123
container-name: my-app
task-def: taskdef.json
```

Deploy:

```bash
oxctl deploy --oxconf
```

For interactive TUI mode (no args in a TTY):

```bash
oxctl
```

---

## Commands

### `oxctl deploy`

Registers a new ECS task definition revision and creates or updates the service.

**Usage:**

```bash
# Via oxconf file
oxctl deploy --oxconf

# Via inline flags
oxctl deploy \
  --cluster my-cluster \
  --service my-service \
  --image 123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:tag \
  --container-name my-app \
  --task-def taskdef.json
```

**Flags:**

| Flag | Type | Description |
|------|------|-------------|
| `--oxconf` | bool | Load config from `./oxconf` (ignores other flags) |
| `--oxconf-path` | string | Path to oxconf file (default: `./oxconf`) |
| `--cluster` | string | ECS cluster name |
| `--service` | string | ECS service name |
| `--image` | string | Full container image URI with tag |
| `--container-name` | string | Container name in task definition to update |
| `--task-def` | string | Path to task definition JSON file |
| `--wait` | bool | Wait for service to reach stable state |
| `--timeout` | int | Stability wait timeout in seconds (default: 300) |

### `oxctl status`

Shows ECS service status: running count, desired count, active deployments, and stability status.

**Usage:**

```bash
# Via oxconf file
oxctl status --oxconf

# Via inline flags
oxctl status \
  --cluster my-cluster \
  --service my-service
```

**Flags:**

| Flag | Type | Description |
|------|------|-------------|
| `--oxconf` | bool | Load cluster/service from `./oxconf` |
| `--oxconf-path` | string | Path to oxconf file (default: `./oxconf`) |
| `--cluster` | string | ECS cluster name |
| `--service` | string | ECS service name |

### Interactive TUI

Run oxctl with no arguments in a terminal (TTY):

```bash
oxctl
```

This launches an interactive Bubble Tea UI with a menu to select deploy or status, a command builder, and output view. The TUI is automatically disabled in CI (detected via common `CI`, `GITHUB_ACTIONS`, `GITLAB_CI`, `AZURE_PIPELINES` environment variables).

---

## Global Flags

Available on all commands:

| Flag | Type | Description |
|------|------|-------------|
| `--json-log` | bool | Output structured JSON logs instead of text |
| `--debug` | bool | Enable verbose debug logging |
| `--dry-run` | bool | Print AWS CLI commands without executing them |

**Examples:**

```bash
# Debug mode to see AWS CLI calls
oxctl deploy --oxconf --debug

# Dry run to preview commands
oxctl deploy --oxconf --dry-run

# JSON logging for log aggregation
oxctl deploy --oxconf --json-log
```

---

## oxconf File Format

YAML configuration file (`oxconf`, no extension). All required fields must be present for first deploy; update deploys only use `cluster`, `service`, `image`, `container-name`, `task-def`, `wait`, and `timeout`.

### Required Fields

| Field | Type | Example | Purpose |
|-------|------|---------|---------|
| `cluster` | string | `bc-frontend-cluster-dev` | ECS cluster name |
| `service` | string | `my-app-dev` | ECS service name |
| `image` | string | `123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:abc123` | Container image URI with tag |
| `container-name` | string | `my-app` | Container name in task definition to update |
| `task-def` | string | `taskdef.json` | Path to task definition JSON file |

### Optional Fields

| Field | Type | Default | Purpose |
|-------|------|---------|---------|
| `wait` | bool | `false` | Wait for service to reach stable state after deploy |
| `timeout` | int | `300` | Stability wait timeout (seconds) |
| `container-port` | int | `3000` | Container port (only used on first service creation) |
| `desired-count` | int | `1` | Number of task instances (only used on first creation) |
| `subnet-ids` | string[] | — | VPC subnet IDs (required for first-time service creation) |
| `security-group-ids` | string[] | — | Security group IDs (required for first-time service creation) |
| `target-group-arn` | string | — | ALB target group ARN (required for first-time service creation) |
| `log-group` | string | `/ecs/<service>` | CloudWatch log group name to ensure exists |

### Full Example

```yaml
cluster: bc-frontend-cluster-dev
service: my-app-dev
image: 123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:abc123
container-name: my-app
task-def: taskdef.json

# Optional: stability check
wait: true
timeout: 300

# Optional: first-time service creation (only needed once)
container-port: 3000
desired-count: 1
subnet-ids:
  - subnet-0abc123def456789
  - subnet-0fed987cba321098
security-group-ids:
  - sg-0abc123def456789
target-group-arn: arn:aws:elasticloadbalancing:us-east-1:123456789:targetgroup/my-app-tg/50dc6c495c0c9a06

# Optional: custom log group
log-group: /ecs/my-app-dev
```

### Important: Existing vs. New Services

If deploying to an **existing service** that was created outside oxctl, you only need the required fields plus `wait`/`timeout`. The `subnet-ids`, `security-group-ids`, and `target-group-arn` fields are only needed on **first-time service creation** with oxctl. Once a service exists, these fields are ignored.

To check if a service exists:

```bash
oxctl status --oxconf
```

If it shows "service not found", include the networking and load balancer fields in oxconf before deploying.

---

## Deploy Logic

On each deploy, oxctl executes these steps in order:

1. **Enable Container Insights** on the ECS cluster for observability
2. **Ensure CloudWatch log group** exists (idempotent — safe to run repeatedly)
3. **Patch task definition** — replaces the container image for the named container
4. **Register new task definition revision** — AWS creates a new numbered revision
5. **Check service status**
   - If **ACTIVE**: Update the service with the new task definition, enable ECS Exec, set health-check grace period to 60s
   - If **MISSING/INACTIVE**: Create a new Fargate service with network config, load balancer, and desired task count
6. **Optionally wait for stability** — polls until desired count equals running count (or timeout)

This design ensures idempotency: running the same deploy twice is safe. If networking fields change on an update deploy, they are ignored (service networking is immutable).

---

## Docker Image

The canonical way to use oxctl in CI/CD is via the Docker image. It includes the oxctl binary and the AWS CLI pre-installed (Amazon Linux 2 base).

### Build Locally

```bash
docker build -t oxctl:dev .
```

### Run in Container

```bash
# Via oxconf
docker run --rm \
  -v $(pwd)/oxconf:/oxconf \
  -v $(pwd)/taskdef.json:/taskdef.json \
  -e AWS_ACCESS_KEY_ID=xxx \
  -e AWS_SECRET_ACCESS_KEY=yyy \
  -e AWS_REGION=us-east-1 \
  oxctl:dev deploy --oxconf

# Via inline flags
docker run --rm \
  -e AWS_ACCESS_KEY_ID=xxx \
  -e AWS_SECRET_ACCESS_KEY=yyy \
  -e AWS_REGION=us-east-1 \
  oxctl:dev deploy \
  --cluster my-cluster \
  --service my-service \
  --image 123456789.dkr.ecr.us-east-1.amazonaws.com/my-app:tag \
  --container-name my-app \
  --task-def /taskdef.json
```

The image is based on `amazon/aws-cli:2.34.38`, so the `aws` CLI is available inside for any pre/post-deploy hooks.

---

## Required AWS Permissions

The IAM role or user running oxctl needs these permissions. Adjust resources as needed:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ecs:UpdateClusterSettings",
        "ecs:DescribeServices",
        "ecs:UpdateService",
        "ecs:CreateService",
        "ecs:RegisterTaskDefinition",
        "ecs:DescribeTaskDefinition"
      ],
      "Resource": [
        "arn:aws:ecs:*:ACCOUNT_ID:cluster/my-cluster",
        "arn:aws:ecs:*:ACCOUNT_ID:service/my-cluster/my-service",
        "arn:aws:ecs:*:ACCOUNT_ID:task-definition/my-task:*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup"
      ],
      "Resource": "arn:aws:logs:*:ACCOUNT_ID:log-group:/ecs/*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ecs:Wait"
      ],
      "Resource": "arn:aws:ecs:*:ACCOUNT_ID:service/my-cluster/my-service"
    },
    {
      "Effect": "Allow",
      "Action": [
        "iam:PassRole"
      ],
      "Resource": [
        "arn:aws:iam::ACCOUNT_ID:role/ecsTaskExecutionRole",
        "arn:aws:iam::ACCOUNT_ID:role/ecsTaskRole"
      ]
    }
  ]
}
```

Key actions:
- `ecs:UpdateClusterSettings` — Enable container insights
- `ecs:DescribeServices` — Check if service exists
- `ecs:UpdateService` — Update existing service
- `ecs:CreateService` — Create new service
- `ecs:RegisterTaskDefinition` — Register task definition
- `ecs:DescribeTaskDefinition` — Read current task definition
- `ecs:Wait` — Wait for service stability
- `logs:CreateLogGroup` — Ensure CloudWatch log group exists
- `iam:PassRole` — Assign execution and task roles

---

## CI/CD Integration

### GitHub Actions

Deploy on push to `main` branch, or on version tags with a manual approval gate for production:

```yaml
name: Deploy

on:
  push:
    branches:
      - main
      - development
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    outputs:
      image-tag: ${{ steps.meta.outputs.tags }}
    steps:
      - uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Amazon ECR
        id: login-ecr
        uses: aws-actions/amazon-ecr-login@v2
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        id: meta
        with:
          context: ./app
          push: true
          tags: |
            ${{ steps.login-ecr.outputs.registry }}/my-app:${{ github.sha }}
            ${{ steps.login-ecr.outputs.registry }}/my-app:latest
          cache-from: type=registry,ref=${{ steps.login-ecr.outputs.registry }}/my-app:buildcache
          cache-to: type=registry,ref=${{ steps.login-ecr.outputs.registry }}/my-app:buildcache,mode=max

  deploy-dev:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main' || github.ref == 'refs/heads/development'
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY }}
          aws-region: us-east-1

      - name: Update oxconf with new image
        run: |
          sed -i "s|image:.*|image: ${{ needs.build.outputs.image-tag }}|" oxconf

      - name: Deploy to dev
        run: |
          docker run --rm \
            -v $(pwd)/oxconf:/oxconf \
            -v $(pwd)/taskdef.json:/taskdef.json \
            -e AWS_ACCESS_KEY_ID=${{ secrets.AWS_ACCESS_KEY_ID }} \
            -e AWS_SECRET_ACCESS_KEY=${{ secrets.AWS_SECRET_ACCESS_KEY }} \
            -e AWS_REGION=us-east-1 \
            ghcr.io/oxgrad/oxctl:latest deploy --oxconf --wait

  deploy-prod:
    needs: build
    runs-on: ubuntu-latest
    if: startsWith(github.ref, 'refs/tags/v')
    environment:
      name: production
      url: https://my-app.example.com
    steps:
      - uses: actions/checkout@v4

      - name: Configure AWS credentials
        uses: aws-actions/configure-aws-credentials@v4
        with:
          aws-access-key-id: ${{ secrets.AWS_ACCESS_KEY_ID_PROD }}
          aws-secret-access-key: ${{ secrets.AWS_SECRET_ACCESS_KEY_PROD }}
          aws-region: us-east-1

      - name: Update oxconf with new image
        run: |
          sed -i "s|image:.*|image: ${{ needs.build.outputs.image-tag }}|" oxconf.prod

      - name: Deploy to prod
        run: |
          docker run --rm \
            -v $(pwd)/oxconf.prod:/oxconf \
            -v $(pwd)/taskdef.json:/taskdef.json \
            -e AWS_ACCESS_KEY_ID=${{ secrets.AWS_ACCESS_KEY_ID_PROD }} \
            -e AWS_SECRET_ACCESS_KEY=${{ secrets.AWS_SECRET_ACCESS_KEY_PROD }} \
            -e AWS_REGION=us-east-1 \
            ghcr.io/oxgrad/oxctl:latest deploy --oxconf --wait --timeout 600
```

**Key points:**

- The build job outputs the image tag; deploy jobs consume it
- Dev deploy runs automatically on `main` or `development` push
- Prod deploy only runs on `v*` tags and requires manual approval via `environment: production`
- The Docker image runs oxctl with the updated oxconf file
- `--wait` polls until the service is stable (optional, useful for catching deploy errors early)

### GitLab CI

Deploy on push to `development` branch automatically; deploy on `v*` tags with manual approval:

```yaml
stages:
  - build
  - docker
  - deploy-dev
  - deploy-prod

variables:
  REGISTRY: $CI_REGISTRY
  IMAGE_NAME: $CI_REGISTRY_IMAGE/my-app
  AWS_REGION: us-east-1

build:
  stage: build
  image: golang:1.26-alpine
  script:
    - go build -o bin/app ./cmd/app
  artifacts:
    paths:
      - bin/app
    expire_in: 1 hour

docker-build:
  stage: docker
  image: docker:latest
  services:
    - docker:dind
  variables:
    DOCKER_DRIVER: overlay2
    DOCKER_TLS_CERTDIR: ""
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - export IMAGE_TAG="${IMAGE_NAME}:${CI_COMMIT_SHA:0:8}"
    - export IMAGE_LATEST="${IMAGE_NAME}:latest"
    - docker build -t $IMAGE_TAG -t $IMAGE_LATEST .
    - docker push $IMAGE_TAG
    - docker push $IMAGE_LATEST
    - echo "IMAGE_TAG=$IMAGE_TAG" > build.env
  artifacts:
    reports:
      dotenv: build.env

deploy-dev:
  stage: deploy-dev
  image: $REGISTRY/$CI_PROJECT_NAMESPACE/oxctl:latest
  environment:
    name: development
    url: https://dev.my-app.example.com
  only:
    - development
  before_script:
    - apk add --no-cache sed
  script:
    - sed -i "s|image:.*|image: $IMAGE_TAG|" oxconf
    - oxctl deploy --oxconf --wait
  variables:
    AWS_DEFAULT_REGION: $AWS_REGION

deploy-prod:
  stage: deploy-prod
  image: $REGISTRY/$CI_PROJECT_NAMESPACE/oxctl:latest
  environment:
    name: production
    url: https://my-app.example.com
  only:
    - tags
  when: manual
  before_script:
    - apk add --no-cache sed
  script:
    - sed -i "s|image:.*|image: $IMAGE_TAG|" oxconf.prod
    - oxctl deploy --oxconf --wait --timeout 600
  variables:
    AWS_DEFAULT_REGION: $AWS_REGION
```

**Key points:**

- The `build.env` artifact from docker job passes `IMAGE_TAG` to deploy jobs
- Dev deploy runs automatically on `development` branch
- Prod deploy only runs on tags and requires manual approval (`when: manual`)
- AWS credentials come from CI/CD variables configured in the project

### Azure Pipelines

Deploy on push to `development` branch; deploy on `v*` tags with environment approval:

```yaml
trigger:
  branches:
    include:
      - main
      - development
    exclude:
      - refs/pull/*
  tags:
    include:
      - v*

pool:
  vmImage: ubuntu-latest

variables:
  acrRegistry: myacr.azurecr.io
  imageRepository: my-app
  containerRegistry: myacr
  awsRegion: us-east-1

stages:
  - stage: Build
    displayName: Build and test
    jobs:
      - job: BuildApp
        displayName: Build Go app
        steps:
          - task: GoTool@0
            inputs:
              version: 1.26
          - script: |
              go build -o $(Build.ArtifactStagingDirectory)/app ./cmd/app
              cp oxconf $(Build.ArtifactStagingDirectory)/
              cp taskdef.json $(Build.ArtifactStagingDirectory)/
            displayName: Build binary

          - task: PublishBuildArtifacts@1
            inputs:
              pathToPublish: $(Build.ArtifactStagingDirectory)
              artifactName: app

  - stage: Docker
    displayName: Build and push Docker image
    dependsOn: Build
    jobs:
      - job: DockerBuild
        displayName: Build and push
        steps:
          - task: Docker@2
            inputs:
              command: build
              repository: $(imageRepository)
              dockerfile: Dockerfile
              containerRegistry: $(containerRegistry)
              tags: |
                $(Build.BuildId)
                latest
            displayName: Build image

          - task: Docker@2
            inputs:
              command: push
              repository: $(imageRepository)
              containerRegistry: $(containerRegistry)
              tags: |
                $(Build.BuildId)
                latest
            displayName: Push image

          - script: echo "##vso[task.setvariable variable=imageTag]$(acrRegistry)/$(imageRepository):$(Build.BuildId)"
            displayName: Set image tag variable

  - stage: DeployDev
    displayName: Deploy to dev
    dependsOn: Docker
    condition: |
      or(
        eq(variables['Build.SourceBranch'], 'refs/heads/main'),
        eq(variables['Build.SourceBranch'], 'refs/heads/development')
      )
    jobs:
      - deployment: DeployDev
        displayName: Deploy to development
        environment: development
        strategy:
          runOnce:
            deploy:
              steps:
                - task: DownloadBuildArtifacts@0
                  inputs:
                    buildType: current
                    downloadType: single
                    artifactName: app
                    downloadPath: $(Pipeline.Workspace)

                - task: AWSShellScript@1
                  inputs:
                    awsCredentials: AWS-Dev
                    regionName: $(awsRegion)
                    scriptType: bash
                    scriptLocation: inlineScript
                    inlineScript: |
                      imageTag=$(imageTag)
                      sed -i "s|image:.*|image: $imageTag|" $(Pipeline.Workspace)/app/oxconf
                      docker run --rm \
                        -v $(Pipeline.Workspace)/app/oxconf:/oxconf \
                        -v $(Pipeline.Workspace)/app/taskdef.json:/taskdef.json \
                        -e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
                        -e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
                        -e AWS_REGION=$(awsRegion) \
                        $(imageTag) deploy --oxconf --wait
                  displayName: Deploy with oxctl

  - stage: DeployProd
    displayName: Deploy to production
    dependsOn: Docker
    condition: startsWith(variables['Build.SourceBranch'], 'refs/tags/v')
    jobs:
      - deployment: DeployProd
        displayName: Deploy to production
        environment: production
        strategy:
          runOnce:
            deploy:
              steps:
                - task: DownloadBuildArtifacts@0
                  inputs:
                    buildType: current
                    downloadType: single
                    artifactName: app
                    downloadPath: $(Pipeline.Workspace)

                - task: AWSShellScript@1
                  inputs:
                    awsCredentials: AWS-Prod
                    regionName: $(awsRegion)
                    scriptType: bash
                    scriptLocation: inlineScript
                    inlineScript: |
                      imageTag=$(imageTag)
                      sed -i "s|image:.*|image: $imageTag|" $(Pipeline.Workspace)/app/oxconf.prod
                      docker run --rm \
                        -v $(Pipeline.Workspace)/app/oxconf.prod:/oxconf \
                        -v $(Pipeline.Workspace)/app/taskdef.json:/taskdef.json \
                        -e AWS_ACCESS_KEY_ID=$(AWS_ACCESS_KEY_ID) \
                        -e AWS_SECRET_ACCESS_KEY=$(AWS_SECRET_ACCESS_KEY) \
                        -e AWS_REGION=$(awsRegion) \
                        $(imageTag) deploy --oxconf --wait --timeout 600
                  displayName: Deploy with oxctl
```

**Key points:**

- Separate environments for dev and prod with approval gates
- Dev deploys automatically on `main` or `development` branch push
- Prod deploys only on `v*` tags and requires environment approval
- AWS credentials are configured via service connections in Azure DevOps
- The Docker image is built in the Docker stage and passed to deploy jobs

---

## Building from Source

### Prerequisites

- Go 1.26 or later

### Build

```bash
# Build the binary
go build -o bin/oxctl ./cmd/oxctl

# Or use make (if available)
make build
```

### Test

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Docker

```bash
# Build the Docker image locally
docker build -t oxctl:dev .

# Test the image
docker run --rm oxctl:dev deploy --help
```

---

## Environment Variables

oxctl respects standard CI environment variables to disable the TUI automatically:

- `CI=true` — Generic CI flag
- `GITHUB_ACTIONS=true` — GitHub Actions
- `GITLAB_CI=true` — GitLab CI
- `AZURE_PIPELINES=true` — Azure Pipelines

If any of these are set, oxctl runs in non-interactive mode even in a TTY.

AWS credentials are read from environment variables or the AWS credential provider chain:

- `AWS_ACCESS_KEY_ID` — AWS access key
- `AWS_SECRET_ACCESS_KEY` — AWS secret key
- `AWS_SESSION_TOKEN` — Optional session token
- `AWS_REGION` — AWS region (default: `us-east-1`)

---

## Troubleshooting

### Service Not Found

If oxctl says "service not found" on the first deploy:

```
Error: creating service: ...
```

Make sure your oxconf file includes the required networking fields for first-time creation:

```yaml
cluster: my-cluster
service: my-service
image: ...
container-name: ...
task-def: ...

# Required for first-time service creation:
subnet-ids:
  - subnet-xxx
  - subnet-yyy
security-group-ids:
  - sg-xxx
target-group-arn: arn:aws:elasticloadbalancing:...
```

Check the section "Existing vs. New Services" above.

### Permission Denied

If you see "AccessDenied" or "UnauthorizedOperation" errors, verify your AWS credentials are set and your IAM role has the required permissions. See "Required AWS Permissions" above.

### Task Definition Not Found

If oxctl fails to patch or load the task definition:

```
Error: loading task definition: open taskdef.json: no such file or directory
```

Verify the `task-def` path in oxconf is correct and relative to where you run oxctl.

### Service Fails to Stabilize

If oxctl times out waiting for service stability:

```
Error: waiting for stability: timeout after 300s
```

The service may have failed to deploy. Check the ECS console or logs:

```bash
# Get service details
aws ecs describe-services \
  --cluster my-cluster \
  --services my-service \
  --region us-east-1

# Get task logs
aws logs tail /ecs/my-service --follow --region us-east-1
```

Increase the timeout with `--timeout 600` or `timeout: 600` in oxconf.

### Dry Run to Preview Commands

To see exactly what AWS CLI commands would be run without executing:

```bash
oxctl deploy --oxconf --dry-run
```

---

## Contributing

Contributions are welcome. Please open an issue or pull request.

### Running Tests

```bash
go test ./...
```

### Code Style

This project follows standard Go conventions. Format code with:

```bash
go fmt ./...
gofmt -s -w .
```

---

## License

See LICENSE file.
