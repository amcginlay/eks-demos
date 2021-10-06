# Configure IAM Role

NOTE: these instructions assume that you are logged on to the AWS console as a sufficiently privileged IAM User.
Alternatively you may have assumed a similarly privileged IAM Role.

Your EKS cluster we be built with an appropriately permissioned EC2 instance in the form of a Cloud9 development environment. The purpose of this section is to ensure that you have an appropriate IAM role, named `Role-EC2-EKSClusterAdmin`, available for the instance to assume. As the creation of the role is a one-time requirement it may first be advisable to check if an instance profile for that role already exists.
```bash
aws iam get-instance-profile --instance-profile-name Role-EC2-EKSClusterAdmin
```

If the above command responds with valid JSON then the role already exists and you can **skip this section**. [Return To Main Menu](/README.md)

If, however, your receive a `NoSuchEntity` error, that's your signal to stay here and continue as follows.

Identify the AWS managed `AdministratorAccess` policy then create the Role-EC2-EKSClusterAdmin role, ensuring both the current user and EC2 instances are able to assume it.
```bash
# NOTE cluster creators should ideally follow these instructions https://eksctl.io/usage/minimum-iam-policies/
principal=$( \
  aws sts get-caller-identity --query '[Arn]' --output text | \
  sed "s/:assumed-role\//:role\//g" | \
  sed "s/:sts::/:iam::/g" | \
  rev | cut -d"/" -f2- | rev \
)
cat > ./Role-EC2-EKSClusterAdmin.trust << EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": "${principal}",
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

admin_policy_arn=$(aws iam list-policies --query "Policies[?PolicyName=='AdministratorAccess'].Arn" --output text)
aws iam create-instance-profile --instance-profile-name Role-EC2-EKSClusterAdmin
aws iam create-role --role-name Role-EC2-EKSClusterAdmin --assume-role-policy-document file://Role-EC2-EKSClusterAdmin.trust --max-session-duration 43200
aws iam add-role-to-instance-profile --instance-profile-name Role-EC2-EKSClusterAdmin --role-name Role-EC2-EKSClusterAdmin
aws iam attach-role-policy --role-name Role-EC2-EKSClusterAdmin --policy-arn ${admin_policy_arn}
```

[Return To Main Menu](/README.md)
