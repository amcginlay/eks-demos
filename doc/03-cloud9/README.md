# Create Cloud9 (EC2) Environment

## Step 1: From CloudShell

As you create your Cloud9 environment, disable the "AWS managed temporary credentials" feature.
Doing so enables the underlying EC2 instance to correctly acknowledge its assigned IAM Role, in this case `Role-EC2-EKSClusterAdmin`.
```bash
cluster_name=dev
subnet_id=$( \
  aws ec2 describe-subnets \
    --filters "Name=availability-zone,Values=${AWS_DEFAULT_REGION}a" "Name=default-for-az,Values=true" \
    --query "Subnets[].SubnetId" \
    --output text \
)
env_id=$( \
  aws cloud9 create-environment-ec2 \
    --name eks-${cluster_name}-$(date +"%Y%m%d%H%M") \
    --instance-type m5.large \
    --image-id amazonlinux-2-x86_64 \
    --subnet-id ${subnet_id} \
    --automatic-stop-time-minutes 1440 \
    --query "environmentId" \
    --output text \
)
echo env_id=${env_id}
sleep 30 && instance_id=$(aws ec2 describe-instances --filters "Name='tag:aws:cloud9:environment',Values='${env_id}'" --query "Reservations[].Instances[0].InstanceId" --output text)
echo instance_id=${instance_id}                                                                          # if blank, wait (sleep) a little longer and repeat previous instruction
aws cloud9 update-environment --environment-id $env_id --managed-credentials-action DISABLE # disable "AWS managed temporary credentials"
aws ec2 associate-iam-instance-profile --instance-id ${instance_id} --iam-instance-profile Name=Role-EC2-EKSClusterAdmin
```

Execute the following command then navigate your browser to the URL it displays before exiting your CloudShell session
```bash
echo -e "\nGo to your new Cloud9 instance at:\nhttps://${AWS_DEFAULT_REGION}.console.aws.amazon.com/cloud9/ide/${env_id}\n"
```

## Step 2: From Cloud9

Once inside the Cloud9 environment, open a terminal session and run the following command to confirm the `Role-EC2-EKSClusterAdmin` IAM role has been assumed:
```bash
aws sts get-caller-identity
```

The standard Cloud9 environment has a small (10gb) root volume.
To ensure you don't exhaust this storage extend the root volume to 30gb.
```bash
region=$(curl --silent http://169.254.169.254/latest/meta-data/placement/region)
instance_id=$(curl --silent http://169.254.169.254/latest/meta-data/instance-id)

volume_id=$(aws ec2 describe-instances \
  --region ${region} \
  --instance-id ${instance_id} \
  --query "Reservations[0].Instances[0].BlockDeviceMappings[0].Ebs.VolumeId" \
  --output text
)

aws ec2 modify-volume \
  --region ${region} \
  --volume-id ${volume_id} \
  --size 30

while [ \
  "$(aws ec2 describe-volumes-modifications \
    --region ${region} \
    --volume-id ${volume_id} \
    --filters Name=modification-state,Values="optimizing","completed" \
    --query "length(VolumesModifications)"\
    --output text)" != "1" ]; do
  sleep 1
done

sudo growpart /dev/nvme0n1 1
sudo xfs_growfs -d /
```

[Return To Main Menu](/README.md)
