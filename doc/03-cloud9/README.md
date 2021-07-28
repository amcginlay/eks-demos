# Create Cloud9 (EC2) Environment

Cloud9 has a feature known as "AWS managed temporary credentials". Before creating our Cloud9 environment we need to disable this feature. Doing so enables the underlying EC2 instance to correctly acknowledge its assigned IAM Role, in this case `Role-EC2-EKSClusterAdmin`. It is not (currently) possible to programatically disable this feature directly from the Cloud9 API however we can assign an inline IAM policy to the currently active principal.

Execute the following from your CloudShell session.
```bash
arn=$(aws sts get-caller-identity --query Arn --output text)
context="user"
if grep -q assumed-role <<< ${arn}; then
  context="role"
fi
principal=$(echo ${arn} | cut -d/ -f2)
aws iam put-${context}-policy \
  --${context}-name ${principal} \
  --policy-name Policy-DisableCloud9Update \
  --policy-document file://<(echo '{"Version": "2012-10-17","Statement": [{"Effect": "Deny","Action": "cloud9:UpdateEnvironment","Resource": "*"}]}')
```

Create your Cloud9 environment from the CloudShell session and associate `Role-EC2-EKSClusterAdmin` with this instance
```bash
cluster_name=dev
env_id=$(aws cloud9 create-environment-ec2 --name c9-eks-${cluster_name} --instance-type m5.large --image-id amazonlinux-2-x86_64 --query "environmentId" --output text)
sleep 25 && instance_id=$(aws ec2 describe-instances --filters "Name='tag:aws:cloud9:environment',Values='${env_id}'" --query "Reservations[].Instances[0].InstanceId" --output text)
echo ${instance_id}                                            # if blank, wait (sleep) a little longer and repeat previous instruction
aws ec2 associate-iam-instance-profile --instance-id ${instance_id} --iam-instance-profile Name=Role-EC2-EKSClusterAdmin
```

Execute the following command then navigate your browser to the URL it displays before exiting/closing your CloudShell session
```bash
echo "https://${AWS_DEFAULT_REGION}.console.aws.amazon.com/cloud9/ide/${env_id}"
```

Once inside the Cloud9 environment, open a terminal session and run the following command to confirm the `Role-EC2-EKSClusterAdmin` IAM role has been assumed:
```bash
aws sts get-caller-identity
```

[Return To Main Menu](/README.md)
