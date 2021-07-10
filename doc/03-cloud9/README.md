# Create Cloud9 (EC2) Environment

Create your Cloud9 environment from the CloudShell session and associate `Role-EC2-EKSClusterAdmin` with this instance
```bash
cluster_name=dev
env_id=$(aws cloud9 create-environment-ec2 --name c9-eks-${cluster_name} --instance-type m5.large --image-id amazonlinux-2-x86_64 --query "environmentId" --output text)
sleep 25 && instance_id=$(aws ec2 describe-instances --filters "Name='tag:aws:cloud9:environment',Values='${env_id}'" --query "Reservations[].Instances[0].InstanceId" --output text)
echo ${instance_id}                                            # if blank, wait (sleep) a little longer and repeat previous instruction
aws ec2 associate-iam-instance-profile --instance-id ${instance_id} --iam-instance-profile Name=Role-EC2-EKSClusterAdmin
```

Execute the following command then navigate your browser to the URL produced before exiting/closing your CloudShell session
```bash
echo "https://${AWS_DEFAULT_REGION}.console.aws.amazon.com/cloud9/ide/${env_id}"
```

[Return To Main Menu](/README.md)
