output "ec2_public_ip" {
  description = "Public IP of EC2 instance"
  value       = aws_eip.k3s.public_ip
}

output "ec2_instance_id" {
  description = "EC2 instance ID"
  value       = aws_instance.k3s.id
}

output "ssh_command" {
  description = "SSH command to connect to EC2"
  value       = "ssh ec2-user@${aws_eip.k3s.public_ip}"
}

output "kubeconfig_command" {
  description = "Command to get kubeconfig"
  value       = "ssh ec2-user@${aws_eip.k3s.public_ip} 'sudo cat /etc/rancher/k3s/k3s.yaml'"
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}

output "subnet_id" {
  description = "Public subnet ID"
  value       = aws_subnet.public.id
}

output "security_group_id" {
  description = "Security group ID"
  value       = aws_security_group.k3s.id
}

output "next_steps" {
  description = "Next steps after deployment"
  value = <<-EOT
    
    ✅ EC2 instance created successfully!
    
    1. Wait 3-5 minutes for k3s installation to complete
    
    2. SSH into the instance:
       ssh ec2-user@${aws_eip.k3s.public_ip}
    
    3. Check k3s status:
       sudo systemctl status k3s
    
    4. Get kubeconfig:
       sudo cat /etc/rancher/k3s/k3s.yaml
    
    5. Configure local kubectl:
       ./scripts/configure-kubectl.sh ${aws_eip.k3s.public_ip}
    
    6. Deploy your services:
       ./scripts/deploy-to-ec2.sh
    
    Public IP: ${aws_eip.k3s.public_ip}
    
  EOT
}