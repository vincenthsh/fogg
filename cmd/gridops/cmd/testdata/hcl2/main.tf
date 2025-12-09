resource "aws_security_group" "example" {
  vpc_id = data.terraform_remote_state.vpc.outputs.vpc_id

  ingress {
    cidr_blocks = [data.terraform_remote_state.networking.outputs.public_subnets]
  }
}

output "debug" {
  value = data.terraform_remote_state.vpc.outputs.name
}
