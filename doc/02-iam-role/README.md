# Create an IAM Role to permit creation of your EKS cluster

NOTE: these instructions have been tested under the assumption that you are logged on to the AWS console as a sufficiently privileged IAM User, *not* an assumed IAM Role. This limitation will be addressed at a later date.

To build our EKS cluster we will use an appropriatelky permissioned EC2 instance in the form of a Cloud9 development environment. The purpose of this section is to ensure that you have an appropriate IAM role available for the instance to assume. As the creation of the role, named `Role-EC2-EKSClusterAdmin`, is a one-time requirement it may first be advisable to check if an instance profile for that role already exists.
```bash
if aws iam get-instance-profile --instance-profile-name Role-EC2-EKSClusterAdminx 2&>1 > /dev/null
then
  echo "Role already exists, skip this section"
else
  echo "Role missing, continue ..."
fi
```

Identify the AWS managed `AdministratorAccess` policy then create the Role-EC2-EKSClusterAdmin role, ensuring both the current user and EC2 instances are able to assume it.
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

[Return To Main Menu](/README.md)
