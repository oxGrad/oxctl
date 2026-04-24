        RegisterTaskDefinition(ctx context.Context, input RegisterTaskInput) (string, error)
        UpdateService(ctx context.Context, input UpdateServiceInput) error
        DescribeService(ctx context.Context, cluster, service string) (ServiceStatus, error)
    }

    type CommandRunner interface {
        Run(ctx context.Context, name string, args ...string) error
        Output(ctx context.Context, name string, args ...string) ([]byte, error)
    }

---

### 8.4 Execution Flow

    CLI → Config Parsing → Orchestrator → ECS Interface → Command Runner → AWS CLI

TUI flow:

    TUI → User Input → Command Builder → CLI Command Output (no execution)

---

## 9. Library Choices

### 9.1 CLI Framework

- Cobra

### 9.2 Logging

- slog (Go standard structured logging)

### 9.3 TUI Framework

- Bubble Tea

### 9.4 Styling

- Lip Gloss

### 9.5 Command Execution

- Go standard library (os/exec)

### 9.6 JSON Handling

- Go standard library (encoding/json)

---

## 10. Deployment Model

### 10.1 Containerized Tool

- Built as Docker image
- Contains:
  - oxctl binary - AWS CLI

        Example:

            docker run oxctl:1.0.0 deploy ...

            ---

            ### 10.2 CI/CD Integration

            - Run container inside pipeline
            - Pass:
              - AWS credentials via environment variables
                - deployment parameters via flags

                ---

                ## 11. Versioning Strategy

                ### Deployer
                - Semantic versioning (v1.0.0)

                ### Application Image
                - Immutable tags (commit SHA)

                ---

                ## 12. Risks & Mitigations

                ### Risk: Over-abstraction
                - Mitigation: keep CLI thin and explicit

                ### Risk: TUI duplication of logic
                - Mitigation: reuse command builder logic from core modules

                ### Risk: Accidental TUI execution in CI
                - Mitigation: detect non-TTY and CI environment

                ### Risk: Debugging difficulty
                - Mitigation: log full AWS CLI commands

                ---

                ## 13. Future Enhancements

                - AWS SDK backend
                - Config file support (YAML/JSON)
                - Deployment hooks (pre/post)
                - Multi-service orchestration
                - Advanced deployment strategies (blue/green, canary)

                ---

                ## 14. Success Metrics

                - Reduced pipeline complexity
                - Faster onboarding for new users
                - Deployment success rate
                - Time to debug failures

                ---

                ## 15. Summary

                oxctl is a focused CLI tool that:
                - wraps AWS CLI
                - keeps deployments stateless
                - provides an optional interactive TUI for human users

                It is designed to be:
                - simple
                - predictable
                - portable across CI/CD systems

    }

# Product Requirements Document (PRD)

## Project Name

oxctl

---

## 1. Overview

oxctl is a lightweight CLI tool written in Go that wraps the AWS CLI to perform deterministic, stateless deployments to Amazon ECS.

The tool is designed as a thin orchestration layer, not a full abstraction over AWS. It prioritizes:

- simplicity
- reproducibility
- CI/CD portability

Additionally, oxctl provides an **interactive TUI mode** for human users when run without arguments, acting as an onboarding and command-generation interface.

---

## 2. Goals

### Primary Goals

- Provide a clean CLI wrapper around AWS CLI
- Enable stateless ECS deployments
- Ensure clear and maintainable project structure
- Optimize for fast iteration and low cognitive overhead
- Be container-friendly for CI/CD usage

### Secondary Goals

- Provide an interactive TUI for discoverability and usability
- Support dry-run execution
- Provide structured logging
- Allow future extension (SDK backend, more commands)

---

## 3. Non-Goals

- Blue/green or advanced deployment strategies (future scope)
- Infrastructure provisioning (handled externally)
- Full ECS abstraction layer
- UI dashboard (web-based)
- Multi-cloud support

---

## 4. Target Users

- DevOps engineers
- Platform engineers
- Backend engineers deploying to ECS

Assumptions:

- Users understand ECS concepts
- Users are comfortable with CLI tools
- Users operate within CI/CD pipelines

---

## 5. Core Use Cases

### 5.1 Standard Deployment Flow

1. Build container image
2. Push to registry
3. Run oxctl to:
    - register new task definition
      - update ECS service
        - optionally wait for stability

        ### 5.2 Interactive Exploration (Human Mode)

        - Run oxctl without arguments
        - Navigate TUI to:
          - learn commands
            - generate valid CLI commands
              - explore examples

              ***

              ## 6. Functional Requirements

              ### 6.1 CLI Commands

              #### deploy

                  oxctl deploy \
                        --cluster <cluster> \
                              --service <service> \
                                    --image <image> \
                                          --container-name <name> \
                                                --task-def <path> \
                                                      --wait \
                                                            --timeout 300s

                                                            #### status

                                                                oxctl status \
                                                                      --cluster <cluster> \
                                                                            --service <service>

                                                                            ---

                                                                            ### 6.2 CLI Behavior

                                                                            - Register new ECS task definition
                                                                            - Update ECS service with new task definition
                                                                            - Optionally wait for deployment stability
                                                                            - Support dry-run mode

                                                                            ---

                                                                            ### 6.3 Interactive TUI Mode

                                                                            #### Trigger Condition
                                                                            - Activated when:
                                                                              - no arguments are provided
                                                                                - running in a terminal (TTY)
                                                                                - Disabled automatically in CI environments

                                                                                #### Purpose
                                                                                - Provide interactive documentation
                                                                                - Help users construct valid commands
                                                                                - Improve onboarding and discoverability

                                                                                #### Features
                                                                                - Menu-based navigation
                                                                                - Command builder for deploy
                                                                                - Example command display
                                                                                - Exit without side effects

                                                                                #### Non-Behavior
                                                                                - Does NOT execute deployments
                                                                                - Does NOT replace CLI functionality

                                                                                ---

                                                                                ### 6.4 Logging

                                                                                - Structured logs (key-value format)
                                                                                - Optional JSON output
                                                                                - Log levels:
                                                                                  - info
                                                                                    - debug
                                                                                      - error

                                                                                      ---

                                                                                      ### 6.5 Dry Run Mode

                                                                                      - Print AWS CLI commands without executing
                                                                                      - Used for debugging and CI validation

                                                                                      ---

                                                                                      ## 7. Non-Functional Requirements

                                                                                      ### 7.1 Performance
                                                                                      - Fast startup (<1s)
                                                                                      - Minimal overhead beyond AWS CLI execution

                                                                                      ### 7.2 Portability
                                                                                      - Must run in Docker container
                                                                                      - No dependency on CI provider features

                                                                                      ### 7.3 Reliability
                                                                                      - Clear error messages
                                                                                      - Fail fast on invalid input

                                                                                      ### 7.4 Security
                                                                                      - Credentials via environment variables
                                                                                      - No secret logging

                                                                                      ---

                                                                                      ## 8. Technical Architecture

                                                                                      ### 8.1 Design Principles

                                                                                      - Thin wrapper over AWS CLI
                                                                                      - Clear separation of concerns
                                                                                      - Minimal abstraction
                                                                                      - Interface-driven for extensibility
                                                                                      - TUI and CLI must share core logic

                                                                                      ---

                                                                                      ### 8.2 Project Structure

                                                                                          cmd/
                                                                                                oxctl/
                                                                                                        main.go

                                                                                                            internal/
                                                                                                                  app/          # orchestration logic
                                                                                                                        ecs/          # ECS-specific domain logic
                                                                                                                              runner/       # command execution layer
                                                                                                                                    config/       # CLI parsing and config
                                                                                                                                          tui/          # interactive TUI implementation
                                                                                                                                                log/          # logging abstraction

                                                                                                                                                    pkg/
                                                                                                                                                          util/         # optional shared utilities

                                                                                                                                                          ---

                                                                                                                                                          ### 8.3 Core Interfaces

                                                                                                                                                              type ECSDeployer interface {
                                                                                                                                                              }
