terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.0"
    }
  }
}

provider "aws" {
  region = "us-east-1"
}

variable "existing_github_oidc_provider_arn" {
  description = "Set this if the AWS account already has the GitHub Actions OIDC provider."
  type        = string
  default     = null
}

resource "aws_ecrpublic_repository" "app" {
  repository_name = "cloud-infrastructure-platform"

  catalog_data {
    description       = "Go Task API and database migration images for the Cloud Infrastructure Platform learning project."
    architectures     = ["x86-64"]
    operating_systems = ["Linux"]
  }

  tags = {
    Project   = "cloud-infrastructure-platform"
    ManagedBy = "Terraform"
  }
}

resource "aws_iam_openid_connect_provider" "github" {
  count = var.existing_github_oidc_provider_arn == null ? 1 : 0

  url            = "https://token.actions.githubusercontent.com"
  client_id_list = ["sts.amazonaws.com"]

  tags = {
    Project   = "cloud-infrastructure-platform"
    ManagedBy = "Terraform"
  }
}

locals {
  github_oidc_provider_arn = var.existing_github_oidc_provider_arn != null ? var.existing_github_oidc_provider_arn : aws_iam_openid_connect_provider.github[0].arn
}

data "aws_iam_policy_document" "assume_from_github_main" {
  statement {
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [local.github_oidc_provider_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:aud"
      values   = ["sts.amazonaws.com"]
    }

    condition {
      test     = "StringEquals"
      variable = "token.actions.githubusercontent.com:sub"
      values   = ["repo:shivamshashank/cloud-infrastructure-platform:ref:refs/heads/main"]
    }
  }
}

resource "aws_iam_role" "github_ecr_public_publish" {
  name               = "cloud-infrastructure-platform-ecr-public-publish"
  assume_role_policy = data.aws_iam_policy_document.assume_from_github_main.json

  tags = {
    Project   = "cloud-infrastructure-platform"
    ManagedBy = "Terraform"
  }
}

data "aws_iam_policy_document" "publish" {
  statement {
    sid = "GetPublicRegistryToken"
    actions = [
      "ecr-public:GetAuthorizationToken",
      "sts:GetServiceBearerToken",
    ]
    resources = ["*"]
  }

  statement {
    sid = "PublishOnlyThisRepository"
    actions = [
      "ecr-public:BatchCheckLayerAvailability",
      "ecr-public:CompleteLayerUpload",
      "ecr-public:DescribeImages",
      "ecr-public:DescribeRepositories",
      "ecr-public:InitiateLayerUpload",
      "ecr-public:PutImage",
      "ecr-public:UploadLayerPart",
    ]
    resources = [aws_ecrpublic_repository.app.arn]
  }
}

resource "aws_iam_role_policy" "publish" {
  name   = "ecr-public-publish"
  role   = aws_iam_role.github_ecr_public_publish.id
  policy = data.aws_iam_policy_document.publish.json
}

output "repository_uri" {
  description = "Public Docker image prefix for API and migration tags."
  value       = aws_ecrpublic_repository.app.repository_uri
}

output "github_ecr_publish_role_arn" {
  description = "Set this as the AWS_ECR_PUBLISH_ROLE_ARN GitHub Actions variable."
  value       = aws_iam_role.github_ecr_public_publish.arn
}
