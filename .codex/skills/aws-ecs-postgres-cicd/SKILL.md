---
name: aws-ecs-postgres-cicd
description: Create or update GitHub Actions CI/CD workflows and CloudFormation templates that deploy Dockerized apps to AWS ECS Fargate with RDS Postgres. Use when scaffolding or modifying `.github/workflows/ci.yml` and `cloudformation/backend.yaml`, wiring GitHub Secrets into env vars, using OIDC with `AWS_ROLE_ARN`, and auto-generating DB name/user/password with Secrets Manager (including excluded characters for DB password).
---

# AWS ECS + Postgres CI/CD

## Overview
Create a CI/CD pipeline that builds and pushes a Docker image to ECR and deploys an ECS Fargate service plus an RDS Postgres database via CloudFormation.

## Quick start
- Copy `assets/ci.yml` to `.github/workflows/ci.yml`.
- Copy `assets/backend.yaml` to `cloudformation/backend.yaml`.
- Replace service names, region, and CPU/memory defaults; prune or rename app-specific secrets and parameters.

## GitHub Actions workflow (ci.yml)
- Keep `permissions: id-token: write` and use `aws-actions/configure-aws-credentials@v4` with `role-to-assume: ${{ secrets.AWS_ROLE_ARN }}`.
- Define all app secrets in GitHub Secrets and map them in the deploy job `env:`; use GitHub `vars` for non-secret tunables.
- Keep the ECR login/build/push and `aws cloudformation deploy` step with `--parameter-overrides` sourced from env.
- Update test/build steps to match the repo's stack (the template uses Go as the example).
- Avoid access keys; rely on the OIDC role.

## CloudFormation template (backend.yaml)
- Keep the `AWS::SecretsManager::Secret` that generates DB credentials automatically.
- Keep `ExcludeCharacters` exactly `'"@/\\?#&%+=:'` to avoid escape errors.
- Do not require user input for DB name/user/password; use a default `DBName` and the generated username/password.
- Build `DATABASE_URL` in the ECS task definition using `{{resolve:secretsmanager:...}}` references.
- Keep Postgres engine config and ECS Fargate + ALB wiring intact unless the architecture changes.

## Outputs
- Emit stack outputs for service URL, database endpoint, and DB credentials ARN for debugging and handoff.
