locals {
    # security group for RDS instance
    description = "Security group for RDS instance"
    ingress_with_cidr_blocks = [
        {
            from_port   = 5432
            to_port     = 5432
            protocol    = "tcp"
            # cidr_blocks = [local.vpc_vpc_cidr_block]
            cidr_blocks = [data.terraform_remote_state.vpc.outputs.vpc_cidr_block]
        },
    ]
    name = "rds-security-group"
    # vpc_id = local.vpc_vpc_id
    vpc_id = data.terraform_remote_state.vpc.outputs.vpc_id
    # RDS instance parametersallocated_storage
    # db_subnet_group_name   = local.vpc_database_subnet_group
    db_subnet_group_name   = data.terraform_remote_state.vpc.outputs.database_subnet_group
    engine                 = "postgres"
    engine_version         = "14"
    family                 = "postgres14" # DB parameter group
    major_engine_version   = "14"         # DB option group
    instance_class         = "db.t3.micro"
    identifier             = "test-rds"
    allocated_storage      = 20
    max_allocated_storage  = 100
    vpc_security_group_ids = [module.security_group.security_group_id]
}
