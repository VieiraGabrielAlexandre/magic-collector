resource "aws_key_pair" "deploy" {
  key_name   = "${var.app_name}-key"
  public_key = var.ssh_public_key

  tags = {
    Name    = "${var.app_name}-key"
    Project = var.app_name
  }
}

resource "aws_security_group" "app" {
  name        = "${var.app_name}-sg"
  description = "Magic Collector - HTTP, HTTPS e SSH"

  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.allowed_ssh_cidr]
  }

  ingress {
    description      = "HTTP IPv4"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
  }

  ingress {
    description      = "HTTP IPv6"
    from_port        = 80
    to_port          = 80
    protocol         = "tcp"
    ipv6_cidr_blocks = ["::/0"]
  }

  ingress {
    description      = "HTTPS IPv4"
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
  }

  ingress {
    description      = "HTTPS IPv6"
    from_port        = 443
    to_port          = 443
    protocol         = "tcp"
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name    = "${var.app_name}-sg"
    Project = var.app_name
  }
}

resource "aws_instance" "app" {
  ami                    = data.aws_ami.ubuntu.id
  instance_type          = var.instance_type
  key_name               = aws_key_pair.deploy.key_name
  vpc_security_group_ids = [aws_security_group.app.id]

  root_block_device {
    volume_size           = var.volume_size_gb
    volume_type           = "gp3"
    delete_on_termination = true
  }

  user_data = file("${path.module}/userdata.sh")

  tags = {
    Name    = var.app_name
    Project = var.app_name
  }
}

resource "aws_eip" "app" {
  instance = aws_instance.app.id
  domain   = "vpc"

  tags = {
    Name    = "${var.app_name}-eip"
    Project = var.app_name
  }
}
