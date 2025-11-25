data "terraform_remote_state" "vpc" {
  backend = "http"

  config = {
    address = "https://grid.example.com/tfstate/11111111-1111-1111-1111-111111111111"
  }
}

data "terraform_remote_state" "networking" {
  backend = "http"

  config = {
    address = "https://grid.example.com/tfstate/22222222-2222-2222-2222-222222222222"
  }
}
