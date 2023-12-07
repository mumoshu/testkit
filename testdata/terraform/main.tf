// This example provisions
// - an EKS cluster
// - an S3 bucket

variable "prefix" {
    type = string
    description = "The prefix to use for all resources in this example"
}

variable "vpc_id" {
    type = string
    description = "The id of the VPC to use for this example"
}

variable "region" {
    type = string
    description = "The region to use for this example"
}

// vpc cidr block
data "aws_vpc" "vpc" {
    id = var.vpc_id
}

resource "aws_s3_bucket" "bucket" {
    bucket = "${var.prefix}-bucket"
}

resource "aws_eks_cluster" "cluster" {
    name = "${var.prefix}-cluster"
    role_arn = aws_iam_role.cluster.arn
    vpc_config {
        subnet_ids = aws_subnet.public[*].id
        security_group_ids = [aws_security_group.cluster.id]
    }
}

resource "aws_iam_role" "cluster" {
    name = "${var.prefix}-cluster"
    assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
                "Service": "eks.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}

resource "aws_security_group" "cluster" {
    name = "${var.prefix}-cluster"
    vpc_id = data.aws_vpc.vpc.id
    ingress {
        from_port = 443
        to_port = 443
        protocol = "tcp"
        cidr_blocks = ["0.0.0.0/0"]
    }
    egress {
        from_port = 0
        to_port = 0
        protocol = "-1"
        cidr_blocks = ["0.0.0.0/0"]
    }
}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_subnet" "public" {
    count = 2
    vpc_id = data.aws_vpc.vpc.id
    cidr_block = "${cidrsubnet(data.aws_vpc.vpc.cidr_block, 4, 10+count.index)}"
    availability_zone = data.aws_availability_zones.available.names[count.index%length(data.aws_availability_zones.available.names)]
}
