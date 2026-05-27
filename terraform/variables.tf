variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "instance_type" {
  description = "EC2 instance type"
  type        = string
  default     = "t3.micro"
}

variable "ssh_public_key" {
  description = "Conteúdo da chave pública SSH (ex: cat ~/.ssh/id_rsa.pub)"
  type        = string
}

variable "allowed_ssh_cidr" {
  description = "CIDR liberado para SSH. Use seu IP: curl ifconfig.me/ip"
  type        = string
  default     = "0.0.0.0/0"
}

variable "app_name" {
  description = "Prefixo usado nos nomes dos recursos AWS"
  type        = string
  default     = "magic-collector"
}

variable "volume_size_gb" {
  description = "Tamanho do disco EBS em GB"
  type        = number
  default     = 20
}
