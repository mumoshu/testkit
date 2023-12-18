// This example provisions
// - an S3 bucket

variable "prefix" {
    type = string
    description = "The prefix to use for all resources in this example"
}

resource "aws_s3_bucket" "bucket" {
    bucket = "${var.prefix}-bucket"
}
