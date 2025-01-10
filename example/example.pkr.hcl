packer {
  required_plugins {
    letscloud = {
      version = ">= 0.1.1"
      source  = "github.com/letscloud-community/letscloud"
    }
  }
}

source "letscloud" "example" {
  api_key              = "YOUR-API-KEY"
  location_slug        = "mia1"
  plan_slug            = "1vcpu-1gb-10ssd"
  image_slug           = "ubuntu-24.04-x86_64"
}

build {
  sources = ["source.letscloud.example"]

  provisioner "shell" {
    inline = [
      "echo Hello World",
    ]
  }

}