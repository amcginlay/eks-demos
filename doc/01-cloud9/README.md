# Create Cloud9 (EC2) Environment

NOTE: these instructions have been tested under the assumption that you are logged on to the AWS console as a sufficiently privileged IAM User, *not* an assumed IAM Role.

Navigate to https://us-west-2.console.aws.amazon.com/cloudshell

Update the AWS CLI.
```bash
sudo mv /usr/local/bin/aws /usr/local/bin/aws.old 2> /dev/null
sudo mv /usr/local/aws-cli /usr/local/aws-cli.old 2> /dev/null
rm -r aws awscliv2.zip 2> /dev/null
curl --silent "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install --update
```

Set a variable for the EKS cluster name.
```bash
cluster_name=dev
```

Check you're not using an assumed IAM Role. These instructions have been tested correct for ARN's which include the word "user".
```bash
aws sts get-caller-identity
```

Identify the AWS managed AdministratorAccess policy then create the Role-EC2-EKSClusterAdmin role, ensuring both the current user and EC2 instances are able to assume it. **You can skip this section if you've already created the Role-EC2-EKSClusterAdmin role**
```bash
# NOTE cluster creators should ideally follow these instructions https://eksctl.io/usage/minimum-iam-policies/
admin_policy_arn=$(aws iam list-policies --query "Policies[?PolicyName=='AdministratorAccess'].Arn" --output text)

cat > ./Role-EC2-EKSClusterAdmin.trust << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "$(aws sts get-caller-identity --query '[Arn]' --output text)",
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
aws iam create-instance-profile --instance-profile-name Role-EC2-EKSClusterAdmin
aws iam create-role --role-name Role-EC2-EKSClusterAdmin --assume-role-policy-document file://Role-EC2-EKSClusterAdmin.trust
aws iam add-role-to-instance-profile --instance-profile-name Role-EC2-EKSClusterAdmin --role-name Role-EC2-EKSClusterAdmin
aws iam attach-role-policy --role-name Role-EC2-EKSClusterAdmin --policy-arn ${admin_policy_arn}
```

Create your Cloud9 environment from the CloudShell session and associate new role with this instance
```bash
env_id=$(aws cloud9 create-environment-ec2 --name c9-eks-${cluster_name} --instance-type m5.large --query "environmentId" --output text)
sleep 20 && instance_id=$(aws ec2 describe-instances --filters "Name='tag:aws:cloud9:environment',Values='${env_id}'" --query "Reservations[].Instances[0].InstanceId" --output text)
echo ${instance_id}                                            # if blank, wait (sleep) a little longer and repeat previous instruction
aws ec2 associate-iam-instance-profile --instance-id ${instance_id} --iam-instance-profile Name=Role-EC2-EKSClusterAdmin
```

Execute the following command then navigate your browser to the URL produced before exiting/closing your CloudShell session
```bash
echo "https://${AWS_DEFAULT_REGION}.console.aws.amazon.com/cloud9/ide/${env_id}"
```

[Return To Main Menu](/README.md)
