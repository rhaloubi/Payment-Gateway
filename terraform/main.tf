terraform {
  required_version = ">= 1.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

# VPC
resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name        = "payment-gateway-vpc"
    Environment = var.environment
    ManagedBy   = "Terraform"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id

  tags = {
    Name        = "payment-gateway-igw"
    Environment = var.environment
  }
}

# Public Subnet
resource "aws_subnet" "public" {
  vpc_id                  = aws_vpc.main.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = true

  tags = {
    Name        = "payment-gateway-public-subnet"
    Environment = var.environment
  }
}

# Route Table for Public Subnet
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }

  tags = {
    Name        = "payment-gateway-public-rt"
    Environment = var.environment
  }
}

# Route Table Association
resource "aws_route_table_association" "public" {
  subnet_id      = aws_subnet.public.id
  route_table_id = aws_route_table.public.id
}

# Security Group for EC2
resource "aws_security_group" "k3s" {
  name        = "payment-gateway-k3s-sg"
  description = "Security group for k3s cluster"
  vpc_id      = aws_vpc.main.id

  # SSH
  ingress {
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = var.allowed_ssh_cidr
    description = "SSH access"
  }

  # k3s API Server
  ingress {
    from_port   = 6443
    to_port     = 6443
    protocol    = "tcp"
    cidr_blocks = var.allowed_ssh_cidr
    description = "k3s API server"
  }

  # HTTP (for Cloudflare Tunnel - localhost only)
  ingress {
    from_port   = 30080
    to_port     = 30080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "NGINX Ingress HTTP"
  }

  # HTTPS (for Cloudflare Tunnel - localhost only)
  ingress {
    from_port   = 30443
    to_port     = 30443
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
    description = "NGINX Ingress HTTPS"
  }

  # Prometheus (restricted)
  ingress {
    from_port   = 30090
    to_port     = 30090
    protocol    = "tcp"
    cidr_blocks = var.allowed_ssh_cidr
    description = "Prometheus"
  }

  # Grafana (restricted)
  ingress {
    from_port   = 30300
    to_port     = 30300
    protocol    = "tcp"
    cidr_blocks = var.allowed_ssh_cidr
    description = "Grafana"
  }

  # Vault (restricted)
  ingress {
    from_port   = 30820
    to_port     = 30820
    protocol    = "tcp"
    cidr_blocks = var.allowed_ssh_cidr
    description = "Vault"
  }

  # All outbound traffic
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "All outbound traffic"
  }

  tags = {
    Name        = "payment-gateway-k3s-sg"
    Environment = var.environment
  }
}

# Key Pair
resource "aws_key_pair" "k3s" {
  key_name   = "payment-gateway-k3s-key"
  public_key = var.ssh_public_key

  tags = {
    Name        = "payment-gateway-k3s-key"
    Environment = var.environment
  }
}

# EC2 Instance for k3s
resource "aws_instance" "k3s" {
  ami                    = data.aws_ami.amazon_linux_2023.id
  instance_type          = var.instance_type
  subnet_id              = aws_subnet.public.id
  vpc_security_group_ids = [aws_security_group.k3s.id]
  key_name               = aws_key_pair.k3s.key_name

  root_block_device {
    volume_size           = 50
    volume_type           = "gp3"
    delete_on_termination = true
    encrypted             = true

    tags = {
      Name        = "payment-gateway-k3s-root"
      Environment = var.environment
    }
  }

  user_data = templatefile("${path.module}/user-data.sh", {
    k3s_version = var.k3s_version
  })

  tags = {
    Name        = "payment-gateway-k3s"
    Environment = var.environment
    Role        = "k3s-server"
  }

  lifecycle {
    ignore_changes = [ami]
  }
  
  credit_specification {
    cpu_credits = "standard"
  }
}

# Elastic IP
resource "aws_eip" "k3s" {
  instance = aws_instance.k3s.id
  domain   = "vpc"

  tags = {
    Name        = "payment-gateway-k3s-eip"
    Environment = var.environment
  }

  depends_on = [aws_internet_gateway.main]
}

# Data sources
data "aws_availability_zones" "available" {
  state = "available"
}

data "aws_ami" "amazon_linux_2023" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64"]
  }

  filter {
    name   = "virtualization-type"
    values = ["hvm"]
  }
}